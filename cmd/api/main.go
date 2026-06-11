package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"order-system/internal/config"
	"order-system/internal/pkg/db"
	"order-system/internal/pkg/logger"
	"order-system/internal/pkg/otel"
	"order-system/internal/pkg/redis"
	"order-system/internal/router"
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

	if err := otel.Init(cfg.OpenTelemetry.ServiceName, cfg.OpenTelemetry.ExporterType, cfg.OpenTelemetry.Endpoint); err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "error", err)
		os.Exit(1)
	}

	database, err := db.NewConnectionWithIdleTime(cfg.Database.DSN(), cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime, cfg.Database.ConnMaxIdleTime)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// AutoMigrate 仅在开发环境执行，生产环境应使用迁移工具
	if cfg.AppEnv == "development" {
		if err := db.Migrate(database); err != nil {
			logger.Error("Failed to migrate database", "error", err)
			os.Exit(1)
		}
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

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return startHTTPServer(ctx, database, cfg.AppPort, cfg.JWT.Secret, redisClient.Client(), cfg.AllowOrigins, cfg.Kafka.Brokers, cfg.Kafka.Enabled, cfg.WsGateway.Enabled, cfg.WsGateway.InstanceID)
	})

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
		logger.Info("Service stopped", "error", err)
	}

	logger.Info("All services stopped gracefully")
}

func startHTTPServer(ctx context.Context, database *gorm.DB, port int, jwtSecret string, redisClient *goredis.Client, allowOrigins []string, kafkaBrokers []string, kafkaEnabled bool, wsEnabled bool, instanceID string) error {
	r := router.NewRouter(ctx, database, jwtSecret, redisClient, allowOrigins, kafkaBrokers, kafkaEnabled, wsEnabled, instanceID)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("HTTP server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Stopping HTTP server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		return fmt.Errorf("HTTP server failed to serve: %w", err)
	}
}
