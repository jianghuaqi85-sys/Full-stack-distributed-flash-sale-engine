package ws

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	ShardCount = 16
)

type Hub struct {
	shards         []*shard
	broadcast      chan *BroadcastMsg
	register       chan *Client
	unregister     chan *Client
	allowedOrigins map[string]bool
	jwtSecret      []byte
	redisClient    *redis.Client
	upgrader       *websocket.Upgrader
	sync.RWMutex
}

type BroadcastMsg struct {
	Message []byte
	RoomID  string
}

type shard struct {
	sync.RWMutex
	clients   map[*Client]bool
	roomIndex map[string]map[*Client]bool
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	roomID   string
}

func NewHub(jwtSecret string, redisClient *redis.Client) *Hub {
	h := &Hub{
		broadcast:   make(chan *BroadcastMsg, 1024),
		register:    make(chan *Client, 1024),
		unregister:  make(chan *Client, 1024),
		shards:      make([]*shard, ShardCount),
		jwtSecret:   []byte(jwtSecret),
		redisClient: redisClient,
		allowedOrigins: map[string]bool{
			"http://localhost:3000":  true,
			"http://localhost:5173": true,
		},
	}
	for i := 0; i < ShardCount; i++ {
		h.shards[i] = &shard{
			clients:   make(map[*Client]bool),
			roomIndex: make(map[string]map[*Client]bool),
		}
	}
	h.upgrader = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			h.RLock()
			defer h.RUnlock()
			return h.allowedOrigins[origin]
		},
	}
	go h.run()
	return h
}

func (h *Hub) SetAllowedOrigins(origins []string) {
	h.Lock()
	defer h.Unlock()
	h.allowedOrigins = make(map[string]bool)
	for _, origin := range origins {
		h.allowedOrigins[origin] = true
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			shard := h.getShard(client.userID)
			shard.Lock()
			shard.clients[client] = true
			if client.roomID != "" {
				if _, ok := shard.roomIndex[client.roomID]; !ok {
					shard.roomIndex[client.roomID] = make(map[*Client]bool)
				}
				shard.roomIndex[client.roomID][client] = true
			}
			shard.Unlock()
		case client := <-h.unregister:
			shard := h.getShard(client.userID)
			shard.Lock()
			if _, ok := shard.clients[client]; ok {
				delete(shard.clients, client)
				if client.roomID != "" {
					if roomClients, ok := shard.roomIndex[client.roomID]; ok {
						delete(roomClients, client)
						if len(roomClients) == 0 {
							delete(shard.roomIndex, client.roomID)
						}
					}
				}
				close(client.send)
			}
			shard.Unlock()
		case msg := <-h.broadcast:
			if msg.RoomID != "" {
				h.broadcastToRoom(msg.RoomID, msg.Message)
			} else {
				h.broadcastToAll(msg.Message)
			}
		}
	}
}

func (h *Hub) broadcastToRoom(roomID string, message []byte) {
	for _, shard := range h.shards {
		var stale []*Client

		shard.RLock()
		if roomClients, ok := shard.roomIndex[roomID]; ok {
			for client := range roomClients {
				select {
				case client.send <- message:
				default:
					stale = append(stale, client)
				}
			}
		}
		shard.RUnlock()

		if len(stale) > 0 {
			shard.Lock()
			for _, client := range stale {
				if _, ok := shard.clients[client]; ok {
					delete(shard.clients, client)
					if roomClients, ok := shard.roomIndex[roomID]; ok {
						delete(roomClients, client)
						if len(roomClients) == 0 {
							delete(shard.roomIndex, roomID)
						}
					}
					close(client.send)
				}
			}
			shard.Unlock()
		}
	}
}

func (h *Hub) broadcastToAll(message []byte) {
	for _, shard := range h.shards {
		var stale []*Client

		shard.RLock()
		for client := range shard.clients {
			select {
			case client.send <- message:
			default:
				stale = append(stale, client)
			}
		}
		shard.RUnlock()

		if len(stale) > 0 {
			shard.Lock()
			for _, client := range stale {
				if _, ok := shard.clients[client]; ok {
					delete(shard.clients, client)
					if client.roomID != "" {
						if roomClients, ok := shard.roomIndex[client.roomID]; ok {
							delete(roomClients, client)
							if len(roomClients) == 0 {
								delete(shard.roomIndex, client.roomID)
							}
						}
					}
					close(client.send)
				}
			}
			shard.Unlock()
		}
	}
}

func (h *Hub) getShard(userID string) *shard {
	hash := sha1.Sum([]byte(userID))
	shardIndex := binary.BigEndian.Uint32(hash[:4]) % ShardCount
	return h.shards[shardIndex]
}

func (h *Hub) ServeWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	tokenString := c.Query("token")
	if tokenString == "" {
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "token required"))
		conn.Close()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid token"))
		conn.Close()
		return
	}

	if h.redisClient != nil {
		if exists, err := h.redisClient.Exists(context.Background(), "blacklist:"+tokenString).Result(); err == nil && exists > 0 {
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "token revoked"))
			conn.Close()
			return
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid token claims"))
		conn.Close()
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "user_id not found in token"))
		conn.Close()
		return
	}

	userID := fmt.Sprintf("%d", uint(userIDFloat))
	roomID := c.Query("room_id")

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		roomID: roomID,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(timeNow().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(timeNow().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			break
		}
		c.hub.broadcast <- &BroadcastMsg{
			Message: message,
			RoomID:  c.roomID,
		}
	}
}

func (c *Client) writePump() {
	ticker := timeNewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(timeNow().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(timeNow().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WSMessage WebSocket 消息格式
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SeckillResultPayload 秒杀结果
type SeckillResultPayload struct {
	OrderID   uint   `json:"order_id,omitempty"`
	UserID    uint   `json:"user_id,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// TicketResultPayload 票务结果
type TicketResultPayload struct {
	TicketID   uint   `json:"ticket_id,omitempty"`
	UserID     uint   `json:"user_id,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	OrderNo    string `json:"order_no,omitempty"`
	TicketType string `json:"ticket_type,omitempty"`
	EventTitle string `json:"event_title,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

// SendToUser 向指定用户发送消息
func (h *Hub) SendToUser(userID string, message []byte) {
	shard := h.getShard(userID)
	shard.RLock()
	defer shard.RUnlock()
	for client := range shard.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				// 客户端缓冲满了，跳过
			}
		}
	}
}

// BroadcastToRoom 向房间广播消息
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.broadcast <- &BroadcastMsg{
		Message: message,
		RoomID:  roomID,
	}
}

var (
	timeNow       = func() time.Time { return time.Now() }
	timeNewTicker = func(d time.Duration) *time.Ticker { return time.NewTicker(d) }
)
