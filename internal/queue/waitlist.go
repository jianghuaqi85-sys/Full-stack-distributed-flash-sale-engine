package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WaitlistManager struct {
	redis *redis.Client
}

type WaitlistEntry struct {
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	Position  int       `json:"position"`
	JoinedAt  time.Time `json:"joined_at"`
	Status    string    `json:"status"` // waiting, notified, converted, expired
}

func NewWaitlistManager(redisClient *redis.Client) *WaitlistManager {
	return &WaitlistManager{redis: redisClient}
}

// JoinWaitlist 加入等候名单
func (m *WaitlistManager) JoinWaitlist(ctx context.Context, eventID, userID string) (*WaitlistEntry, error) {
	waitlistKey := fmt.Sprintf("waitlist:event:%s", eventID)
	userKey := fmt.Sprintf("waitlist:user:%s:%s", eventID, userID)

	// 检查是否已在等候名单中
	_, err := m.redis.Get(ctx, userKey).Result()
	if err == nil {
		return nil, fmt.Errorf("您已在等候名单中")
	}

	// 检查是否有库存
	stockKey := fmt.Sprintf("seckill:stock:%s", eventID)
	stock, err := m.redis.HLen(ctx, stockKey).Result()
	if err == nil && stock > 0 {
		return nil, fmt.Errorf("活动有库存，请直接购票")
	}

	// 获取当前位置
	position, err := m.redis.LLen(ctx, waitlistKey).Result()
	if err != nil {
		return nil, fmt.Errorf("加入等候名单失败: %w", err)
	}

	// 添加到等候名单
	entry := &WaitlistEntry{
		UserID:   userID,
		EventID:  eventID,
		Position: int(position) + 1,
		JoinedAt: time.Now(),
		Status:   "waiting",
	}

	if err := m.redis.RPush(ctx, waitlistKey, userID).Err(); err != nil {
		return nil, fmt.Errorf("加入等候名单失败: %w", err)
	}

	// 记录用户信息
	m.redis.Set(ctx, userKey, entry, 7*24*time.Hour) // 7天过期

	return entry, nil
}

// GetWaitlistPosition 获取等候名单位置
func (m *WaitlistManager) GetWaitlistPosition(ctx context.Context, eventID, userID string) (*WaitlistEntry, error) {
	userKey := fmt.Sprintf("waitlist:user:%s:%s", eventID, userID)

	entryStr, err := m.redis.Get(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("您不在等候名单中")
	}

	// 简化处理，返回基本信息
	_ = entryStr

	waitlistKey := fmt.Sprintf("waitlist:event:%s", eventID)
	users, err := m.redis.LRange(ctx, waitlistKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取位置失败: %w", err)
	}

	position := -1
	for i, u := range users {
		if u == userID {
			position = i
			break
		}
	}

	if position == -1 {
		return nil, fmt.Errorf("您不在等候名单中")
	}

	return &WaitlistEntry{
		UserID:   userID,
		EventID:  eventID,
		Position: position + 1,
		Status:   "waiting",
	}, nil
}

// NotifyNext 通知下一位用户
func (m *WaitlistManager) NotifyNext(ctx context.Context, eventID string) (string, error) {
	waitlistKey := fmt.Sprintf("waitlist:event:%s", eventID)

	// 从等候名单头部取出一个用户
	userID, err := m.redis.LPop(ctx, waitlistKey).Result()
	if err != nil {
		return "", fmt.Errorf("等候名单为空")
	}

	// 更新用户状态
	userKey := fmt.Sprintf("waitlist:user:%s:%s", eventID, userID)
	m.redis.Set(ctx, userKey, "notified", 24*time.Hour)

	return userID, nil
}

// LeaveWaitlist 离开等候名单
func (m *WaitlistManager) LeaveWaitlist(ctx context.Context, eventID, userID string) error {
	waitlistKey := fmt.Sprintf("waitlist:event:%s", eventID)
	userKey := fmt.Sprintf("waitlist:user:%s:%s", eventID, userID)

	m.redis.LRem(ctx, waitlistKey, 0, userID)
	m.redis.Del(ctx, userKey)

	return nil
}

// GetWaitlistLength 获取等候名单长度
func (m *WaitlistManager) GetWaitlistLength(ctx context.Context, eventID string) (int64, error) {
	waitlistKey := fmt.Sprintf("waitlist:event:%s", eventID)
	return m.redis.LLen(ctx, waitlistKey).Result()
}
