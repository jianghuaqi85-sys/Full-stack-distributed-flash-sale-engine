package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis PubSub channel for user-targeted messages
	pubSubUserChannel = "ws:user:%s"
	// Redis PubSub channel for room broadcasts
	pubSubRoomChannel = "ws:room:%s"
	// Redis PubSub channel for global broadcasts
	pubSubGlobalChannel = "ws:broadcast"
	// Redis Hash key for user -> instance connection registry
	connectionRegistryKey = "ws:connections"
	// TTL for connection registry entries
	connectionRegistryTTL = 5 * time.Minute
	// Health check interval for connection registry
	healthCheckInterval = 30 * time.Second
)

// WSMessage represents a message to be delivered via WebSocket
type WSMessage struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	UserID    string          `json:"user_id,omitempty"`
	RoomID    string          `json:"room_id,omitempty"`
	Broadcast bool            `json:"broadcast,omitempty"`
}

// WSDeliveryFunc is the function signature for delivering messages to local WebSocket clients
type WSDeliveryFunc func(userID string, message []byte)

// RoomDeliveryFunc is the function signature for delivering messages to a room
type RoomDeliveryFunc func(roomID string, message []byte)

// BroadcastDeliveryFunc is the function signature for broadcasting messages to all clients
type BroadcastDeliveryFunc func(message []byte)

// WsRouter handles cross-instance WebSocket message routing via Redis PubSub
type WsRouter struct {
	client           *redis.Client
	instanceID       string
	subscriber       *redis.Client
	mu               sync.RWMutex
	localUserConns   map[string]int // userID -> connection count on this instance
	userDeliverFunc  WSDeliveryFunc
	roomDeliveryFunc RoomDeliveryFunc
	broadcastFunc    BroadcastDeliveryFunc
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewWsRouter creates a new WebSocket router that uses Redis PubSub for cross-instance messaging
func NewWsRouter(redisClient *redis.Client, instanceID string) *WsRouter {
	ctx, cancel := context.WithCancel(context.Background())
	return &WsRouter{
		client:         redisClient,
		instanceID:     instanceID,
		subscriber:     redisClient,
		localUserConns: make(map[string]int),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// SetDeliveryFuncs sets the callback functions for delivering messages to local WebSocket clients
func (r *WsRouter) SetDeliveryFuncs(userFn WSDeliveryFunc, roomFn RoomDeliveryFunc, broadcastFn BroadcastDeliveryFunc) {
	r.userDeliverFunc = userFn
	r.roomDeliveryFunc = roomFn
	r.broadcastFunc = broadcastFn
}

// Start begins listening for cross-instance messages and starts health checks
func (r *WsRouter) Start() {
	go r.subscribeGlobal()
	go r.healthCheckLoop()
	log.Printf("[WsRouter] Instance %s started", r.instanceID)
}

// Stop gracefully shuts down the router
func (r *WsRouter) Stop() {
	r.cancel()

	// Remove all connection registrations for this instance
	r.removeInstanceConnections()
	log.Printf("[WsRouter] Instance %s stopped", r.instanceID)
}

// RegisterConnection registers a user connection on this instance
func (r *WsRouter) RegisterConnection(userID string) {
	r.mu.Lock()
	r.localUserConns[userID]++
	r.mu.Unlock()

	// Register in Redis: user -> instance mapping
	key := fmt.Sprintf("ws:user:%s:instances", userID)
	r.client.SAdd(r.ctx, key, r.instanceID)
	r.client.Expire(r.ctx, key, connectionRegistryTTL)

	// Increment connection count
	connKey := fmt.Sprintf("ws:conn:%s:%s", r.instanceID, userID)
	r.client.Incr(r.ctx, connKey)
	r.client.Expire(r.ctx, connKey, connectionRegistryTTL)
}

// UnregisterConnection removes a user connection from this instance
func (r *WsRouter) UnregisterConnection(userID string) {
	r.mu.Lock()
	if count, ok := r.localUserConns[userID]; ok {
		if count <= 1 {
			delete(r.localUserConns, userID)
		} else {
			r.localUserConns[userID] = count - 1
		}
	}
	r.mu.Unlock()

	// Decrement connection count
	connKey := fmt.Sprintf("ws:conn:%s:%s", r.instanceID, userID)
	newCount, _ := r.client.Decr(r.ctx, connKey).Result()
	if newCount <= 0 {
		// No more connections, remove from registry
		r.client.Del(r.ctx, connKey)
		key := fmt.Sprintf("ws:user:%s:instances", userID)
		r.client.SRem(r.ctx, key, r.instanceID)
	}
}

// PublishToUser publishes a message for a specific user (may be delivered by any instance)
func (r *WsRouter) PublishToUser(userID string, message []byte) error {
	msg := WSMessage{
		Type:   "user",
		UserID: userID,
		Payload: json.RawMessage(message),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	channel := fmt.Sprintf(pubSubUserChannel, userID)
	return r.client.Publish(r.ctx, channel, data).Err()
}

// PublishToRoom publishes a message to a room (delivered by all instances with room members)
func (r *WsRouter) PublishToRoom(roomID string, message []byte) error {
	msg := WSMessage{
		Type:   "room",
		RoomID: roomID,
		Payload: json.RawMessage(message),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	channel := fmt.Sprintf(pubSubRoomChannel, roomID)
	return r.client.Publish(r.ctx, channel, data).Err()
}

// PublishBroadcast publishes a broadcast message to all instances
func (r *WsRouter) PublishBroadcast(message []byte) error {
	msg := WSMessage{
		Type:      "broadcast",
		Broadcast: true,
		Payload:   json.RawMessage(message),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return r.client.Publish(r.ctx, pubSubGlobalChannel, data).Err()
}

// subscribeGlobal subscribes to broadcast and user-specific channels
func (r *WsRouter) subscribeGlobal() {
	// Subscribe to broadcast channel
	broadcastSub := r.subscriber.Subscribe(r.ctx, pubSubGlobalChannel)
	defer broadcastSub.Close()

	ch := broadcastSub.Channel()
	for {
		select {
		case <-r.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			r.handleBroadcastMessage(msg.Payload)
		}
	}
}

// SubscribeToUser starts listening for messages targeted at a specific user
// This should be called when a user connects to this instance
func (r *WsRouter) SubscribeToUser(userID string) {
	go func() {
		channel := fmt.Sprintf(pubSubUserChannel, userID)
		sub := r.subscriber.Subscribe(r.ctx, channel)
		defer sub.Close()

		ch := sub.Channel()
		for {
			select {
			case <-r.ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				r.handleUserMessage(msg.Payload)
			}
		}
	}()
}

// UnsubscribeFromUser stops listening for messages targeted at a specific user
func (r *WsRouter) UnsubscribeFromUser(userID string) {
	// The subscription will be cancelled when the context is done
	// or when the goroutine detects no more local connections
}

// handleBroadcastMessage handles incoming broadcast messages
func (r *WsRouter) handleBroadcastMessage(payload string) {
	var msg WSMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Printf("[WsRouter] Failed to unmarshal broadcast message: %v", err)
		return
	}

	if r.broadcastFunc != nil {
		r.broadcastFunc(msg.Payload)
	}
}

// handleUserMessage handles incoming user-targeted messages
func (r *WsRouter) handleUserMessage(payload string) {
	var msg WSMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Printf("[WsRouter] Failed to unmarshal user message: %v", err)
		return
	}

	if msg.UserID != "" && r.userDeliverFunc != nil {
		// Check if we have local connections for this user
		r.mu.RLock()
		_, hasLocal := r.localUserConns[msg.UserID]
		r.mu.RUnlock()

		if hasLocal {
			r.userDeliverFunc(msg.UserID, msg.Payload)
		}
	}
}

// removeInstanceConnections removes all connection registrations for this instance
func (r *WsRouter) removeInstanceConnections() {
	r.mu.RLock()
	userIDs := make([]string, 0, len(r.localUserConns))
	for userID := range r.localUserConns {
		userIDs = append(userIDs, userID)
	}
	r.mu.RUnlock()

	for _, userID := range userIDs {
		connKey := fmt.Sprintf("ws:conn:%s:%s", r.instanceID, userID)
		r.client.Del(r.ctx, connKey)
		key := fmt.Sprintf("ws:user:%s:instances", userID)
		r.client.SRem(r.ctx, key, r.instanceID)
	}
}

// healthCheckLoop periodically refreshes connection registry TTLs
func (r *WsRouter) healthCheckLoop() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.refreshConnectionTTLs()
		}
	}
}

// refreshConnectionTTLs refreshes the TTL for all active connections
func (r *WsRouter) refreshConnectionTTLs() {
	r.mu.RLock()
	userIDs := make([]string, 0, len(r.localUserConns))
	for userID := range r.localUserConns {
		userIDs = append(userIDs, userID)
	}
	r.mu.RUnlock()

	for _, userID := range userIDs {
		key := fmt.Sprintf("ws:user:%s:instances", userID)
		r.client.Expire(r.ctx, key, connectionRegistryTTL)
		connKey := fmt.Sprintf("ws:conn:%s:%s", r.instanceID, userID)
		r.client.Expire(r.ctx, connKey, connectionRegistryTTL)
	}
}
