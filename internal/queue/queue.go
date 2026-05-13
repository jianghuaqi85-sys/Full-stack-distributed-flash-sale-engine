package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type QueueManager struct {
	redis    *redis.Client
	queues   map[string]*EventQueue
	mu       sync.RWMutex
}

type EventQueue struct {
	EventID    string
	QueueKey   string
	Processing chan string
	mu         sync.RWMutex
}

type QueuePosition struct {
	Position    int       `json:"position"`
	TotalAhead  int       `json:"total_ahead"`
	EstimatedWait time.Duration `json:"estimated_wait"`
	Status      string    `json:"status"` // waiting, ready, expired
	JoinedAt    time.Time `json:"joined_at"`
}

var (
	globalManager *QueueManager
	once          sync.Once
)

func NewQueueManager(redisClient *redis.Client) *QueueManager {
	once.Do(func() {
		globalManager = &QueueManager{
			redis:  redisClient,
			queues: make(map[string]*EventQueue),
		}
	})
	return globalManager
}

func GetQueueManager() *QueueManager {
	return globalManager
}

// JoinQueue 加入排队队列
func (m *QueueManager) JoinQueue(ctx context.Context, eventID, userID string) (*QueuePosition, error) {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)
	userKey := fmt.Sprintf("queue:user:%s:%s", eventID, userID)

	// 检查用户是否已在队列中
	_, err := m.redis.Get(ctx, userKey).Result()
	if err == nil {
		return nil, fmt.Errorf("您已在排队中，请勿重复提交")
	}

	// 获取当前位置（队列长度）
	position, err := m.redis.LLen(ctx, queueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("加入队列失败: %w", err)
	}

	// 将用户添加到队列末尾
	joinedAt := time.Now().Unix()
	if err := m.redis.RPush(ctx, queueKey, userID).Err(); err != nil {
		return nil, fmt.Errorf("加入队列失败: %w", err)
	}

	// 记录用户入队时间
	m.redis.Set(ctx, userKey, joinedAt, 30*time.Minute)

	// 计算预估等待时间（假设每个用户处理需要 3 秒）
	estimatedWait := time.Duration(position) * 3 * time.Second

	return &QueuePosition{
		Position:      int(position) + 1,
		TotalAhead:    int(position),
		EstimatedWait: estimatedWait,
		Status:        "waiting",
		JoinedAt:      time.Unix(joinedAt, 0),
	}, nil
}

// GetPosition 获取用户在队列中的位置
func (m *QueueManager) GetPosition(ctx context.Context, eventID, userID string) (*QueuePosition, error) {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)
	userKey := fmt.Sprintf("queue:user:%s:%s", eventID, userID)

	// 检查用户是否在队列中
	exists, err := m.redis.Exists(ctx, userKey).Result()
	if err != nil || exists == 0 {
		return nil, fmt.Errorf("您不在队列中")
	}

	// 获取队列中的所有用户
	users, err := m.redis.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取队列位置失败: %w", err)
	}

	// 查找用户位置
	position := -1
	for i, u := range users {
		if u == userID {
			position = i
			break
		}
	}

	if position == -1 {
		return nil, fmt.Errorf("您不在队列中")
	}

	// 获取入队时间
	joinedAtStr, _ := m.redis.Get(ctx, userKey).Result()
	var joinedAt time.Time
	if joinedAtStr != "" {
		// 简化处理，使用当前时间
		joinedAt = time.Now()
	}

	estimatedWait := time.Duration(position) * 3 * time.Second

	return &QueuePosition{
		Position:      position + 1,
		TotalAhead:    position,
		EstimatedWait: estimatedWait,
		Status:        "waiting",
		JoinedAt:      joinedAt,
	}, nil
}

// ProcessNext 处理队列中的下一个用户
func (m *QueueManager) ProcessNext(ctx context.Context, eventID string) (string, error) {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)

	// 从队列头部取出一个用户
	userID, err := m.redis.LPop(ctx, queueKey).Result()
	if err != nil {
		return "", fmt.Errorf("队列为空")
	}

	// 删除用户的排队记录
	userKey := fmt.Sprintf("queue:user:%s:%s", eventID, userID)
	m.redis.Del(ctx, userKey)

	return userID, nil
}

// LeaveQueue 离开队列
func (m *QueueManager) LeaveQueue(ctx context.Context, eventID, userID string) error {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)
	userKey := fmt.Sprintf("queue:user:%s:%s", eventID, userID)

	// 从队列中移除用户
	m.redis.LRem(ctx, queueKey, 0, userID)
	m.redis.Del(ctx, userKey)

	return nil
}

// GetQueueLength 获取队列长度
func (m *QueueManager) GetQueueLength(ctx context.Context, eventID string) (int64, error) {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)
	return m.redis.LLen(ctx, queueKey).Result()
}

// ClearQueue 清空队列
func (m *QueueManager) ClearQueue(ctx context.Context, eventID string) error {
	queueKey := fmt.Sprintf("queue:event:%s", eventID)
	return m.redis.Del(ctx, queueKey).Err()
}
