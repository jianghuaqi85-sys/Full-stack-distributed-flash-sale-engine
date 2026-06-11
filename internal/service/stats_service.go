package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

const (
	dashboardCacheTTL = 30 * time.Second
	salesTrendCacheTTL = 60 * time.Second
	ticketTypeStatsCacheTTL = 60 * time.Second
)

type StatsService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

// NewStatsServiceWithRedis 创建带 Redis 缓存的统计服务
func NewStatsServiceWithRedis(db *gorm.DB, redisClient *redis.Client) *StatsService {
	return &StatsService{db: db, redis: redisClient}
}

// InvalidateStatsCache 主动失效统计缓存（活动发布/票务状态变更时调用）
func (s *StatsService) InvalidateStatsCache() {
	if s.redis == nil {
		return
	}
	ctx := context.Background()
	s.redis.Del(ctx, "stats:dashboard")
	s.redis.Del(ctx, "stats:ticket_types")
	// Sales trend 按日期 key，不需要精确删除
}

type DashboardStats struct {
	TotalEvents    int64   `json:"total_events"`
	ActiveEvents   int64   `json:"active_events"`
	TotalTickets   int64   `json:"total_tickets"`
	SoldTickets    int64   `json:"sold_tickets"`
	ReservedTickets int64  `json:"reserved_tickets"`
	TotalRevenue   float64 `json:"total_revenue"`
	TodaySales     int64   `json:"today_sales"`
	TodayRevenue   float64 `json:"today_revenue"`
}

type SalesTrend struct {
	Date   string  `json:"date"`
	Count  int64   `json:"count"`
	Revenue float64 `json:"revenue"`
}

type TicketTypeStats struct {
	TicketTypeID   uint    `json:"ticket_type_id"`
	TicketTypeName string  `json:"ticket_type_name"`
	EventTitle     string  `json:"event_title"`
	SoldCount      int64   `json:"sold_count"`
	Revenue        float64 `json:"revenue"`
}

type ConversionFunnel struct {
	PageViews    int64 `json:"page_views"`
	AddToCart    int64 `json:"add_to_cart"`
	Reserved     int64 `json:"reserved"`
	Paid         int64 `json:"paid"`
	Used         int64 `json:"used"`
}

func (s *StatsService) GetDashboardStats() (*DashboardStats, error) {
	ctx := context.Background()

	// 尝试从缓存读取
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, "stats:dashboard").Result()
		if err == nil && cached != "" {
			var stats DashboardStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return &stats, nil
			}
		}
	}

	stats := &DashboardStats{}

	// 活动统计
	s.db.Model(&db.Event{}).Count(&stats.TotalEvents)
	s.db.Model(&db.Event{}).Where("status = ?", "on_sale").Count(&stats.ActiveEvents)

	// 票务统计
	s.db.Model(&db.Ticket{}).Count(&stats.TotalTickets)
	s.db.Model(&db.Ticket{}).Where("status = ?", "paid").Count(&stats.SoldTickets)
	s.db.Model(&db.Ticket{}).Where("status = ?", "reserved").Count(&stats.ReservedTickets)

	// 收入统计
	s.db.Model(&db.Ticket{}).Where("status = ?", "paid").Select("COALESCE(SUM(total_price), 0)").Scan(&stats.TotalRevenue)

	// 今日统计
	today := time.Now().Format("2006-01-02")
	s.db.Model(&db.Ticket{}).Where("DATE(created_at) = ? AND status = ?", today, "paid").Count(&stats.TodaySales)
	s.db.Model(&db.Ticket{}).Where("DATE(created_at) = ? AND status = ?", today, "paid").Select("COALESCE(SUM(total_price), 0)").Scan(&stats.TodayRevenue)

	// 写入缓存
	if s.redis != nil {
		if data, err := json.Marshal(stats); err == nil {
			s.redis.Set(ctx, "stats:dashboard", string(data), dashboardCacheTTL)
		}
	}

	return stats, nil
}

func (s *StatsService) GetSalesTrend(days int) ([]SalesTrend, error) {
	ctx := context.Background()
	cacheKey := "stats:sales_trend:" + time.Now().Format("2006-01-02")

	// 尝试从缓存读取
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var trends []SalesTrend
			if json.Unmarshal([]byte(cached), &trends) == nil {
				return trends, nil
			}
		}
	}

	var trends []SalesTrend

	startDate := time.Now().AddDate(0, 0, -(days-1)).Format("2006-01-02")

	type dailyStat struct {
		Date    string
		Count   int64
		Revenue float64
	}
	var stats []dailyStat

	s.db.Model(&db.Ticket{}).
		Select("DATE(created_at) as date, COUNT(*) as count, COALESCE(SUM(total_price), 0) as revenue").
		Where("DATE(created_at) >= ? AND status = ?", startDate, "paid").
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&stats)

	// 构建完整日期序列，补零
	statMap := make(map[string]dailyStat)
	for _, s := range stats {
		statMap[s.Date] = s
	}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if st, ok := statMap[date]; ok {
			trends = append(trends, SalesTrend{
				Date:    st.Date,
				Count:   st.Count,
				Revenue: st.Revenue,
			})
		} else {
			trends = append(trends, SalesTrend{
				Date:    date,
				Count:   0,
				Revenue: 0,
			})
		}
	}

	// 写入缓存
	if s.redis != nil {
		if data, err := json.Marshal(trends); err == nil {
			s.redis.Set(ctx, cacheKey, string(data), salesTrendCacheTTL)
		}
	}

	return trends, nil
}

func (s *StatsService) GetTicketTypeStats() ([]TicketTypeStats, error) {
	ctx := context.Background()
	cacheKey := "stats:ticket_types"

	// 尝试从缓存读取
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var stats []TicketTypeStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return stats, nil
			}
		}
	}

	stats := make([]TicketTypeStats, 0)

	s.db.Model(&db.Ticket{}).
		Select(`
			tickets.ticket_type_id,
			ticket_types.name as ticket_type_name,
			events.title as event_title,
			COUNT(*) as sold_count,
			COALESCE(SUM(tickets.total_price), 0) as revenue
		`).
		Joins("JOIN ticket_types ON tickets.ticket_type_id = ticket_types.id").
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("tickets.status = ?", "paid").
		Group("tickets.ticket_type_id, ticket_types.name, events.title").
		Order("sold_count DESC").
		Scan(&stats)

	// 写入缓存
	if s.redis != nil {
		if data, err := json.Marshal(stats); err == nil {
			s.redis.Set(ctx, cacheKey, string(data), ticketTypeStatsCacheTTL)
		}
	}

	return stats, nil
}

func (s *StatsService) GetConversionFunnel(eventID uint) (*ConversionFunnel, error) {
	funnel := &ConversionFunnel{}

	// 总票务数（所有状态，代表下单意向）
	s.db.Model(&db.Ticket{}).Where("event_id = ?", eventID).Count(&funnel.PageViews)

	// 已预定（reserved 状态，代表已下单未支付）
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "reserved").Count(&funnel.AddToCart)

	// 已支付（paid 状态）
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "paid").Count(&funnel.Reserved)

	// 已使用（used 状态）
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "used").Count(&funnel.Paid)

	// 已取消/过期
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status IN ?", eventID, []string{"cancelled", "expired"}).Count(&funnel.Used)

	return funnel, nil
}
