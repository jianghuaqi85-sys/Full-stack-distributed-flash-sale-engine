package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"order-system/internal/config"
	"order-system/internal/gateway"
	"order-system/internal/pkg/logger"
	"order-system/internal/pkg/otel"
	"order-system/internal/pkg/redis"
	"order-system/internal/pkg/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	if err := cfg.Validate(); err != nil {
		logger.Error("Config validation failed", "error", err)
		os.Exit(1)
	}

	if err := otel.Init(cfg.OpenTelemetry.ServiceName+"-ws-gateway", cfg.OpenTelemetry.ExporterType, cfg.OpenTelemetry.Endpoint); err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "error", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := redisClient.Ping(ctx); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Generate unique instance ID for this gateway
	instanceID := fmt.Sprintf("ws-gateway-%s-%s", cfg.AppEnv, uuid.New().String()[:8])

	// Create WebSocket Hub
	hub := ws.NewHub(cfg.JWT.Secret, redisClient.Client())

	// Set allowed origins from config
	if len(cfg.AllowOrigins) > 0 {
		hub.SetAllowedOrigins(cfg.AllowOrigins)
	}

	// Create and configure the cross-instance router
	wsRouter := gateway.NewWsRouter(redisClient.Client(), instanceID)
	wsRouter.SetDeliveryFuncs(
		func(userID string, message []byte) {
			hub.SendToUserLocal(userID, message)
		},
		func(roomID string, message []byte) {
			hub.BroadcastToRoomLocal(roomID, message)
		},
		func(message []byte) {
			hub.BroadcastToAllLocal(message)
		},
	)

	// Set the router on the hub
	hub.SetWsRouter(wsRouter)

	// Start the router (Redis PubSub listener + health checks)
	wsRouter.Start()

	g, ctx := errgroup.WithContext(ctx)

	// WebSocket HTTP server
	wsPort := 8081 // Different port from main API
	g.Go(func() error {
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			// Use gin context adapter
			hub.ServeHTTP(w, r)
		})
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"ws-gateway","instance":"` + instanceID + `"}`))
		})

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", wsPort),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		logger.Info("WebSocket Gateway listening", "port", wsPort, "instance", instanceID)

		errCh := make(chan error, 1)
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
			close(errCh)
		}()

		select {
		case <-ctx.Done():
			logger.Info("Stopping WebSocket Gateway...")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			wsRouter.Stop()
			return srv.Shutdown(shutdownCtx)
		case err := <-errCh:
			return fmt.Errorf("WebSocket Gateway failed: %w", err)
		}
	})

	// Signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigChan:
			logger.Warn("Received shutdown signal", "signal", sig)
			cancel()
			return fmt.Errorf("received signal: %v", sig)
		}
	})

	if err := g.Wait(); err != nil {
		logger.Info("WebSocket Gateway stopped", "error", err)
	}
	logger.Info("WebSocket Gateway exited gracefully")
}
