package service

import (
	"time"

	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
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

	return stats, nil
}

func (s *StatsService) GetSalesTrend(days int) ([]SalesTrend, error) {
	var trends []SalesTrend

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		var count int64
		var revenue float64

		s.db.Model(&db.Ticket{}).
			Where("DATE(created_at) = ? AND status = ?", date, "paid").
			Count(&count)

		s.db.Model(&db.Ticket{}).
			Where("DATE(created_at) = ? AND status = ?", date, "paid").
			Select("COALESCE(SUM(total_price), 0)").
			Scan(&revenue)

		trends = append(trends, SalesTrend{
			Date:    date,
			Count:   count,
			Revenue: revenue,
		})
	}

	return trends, nil
}

func (s *StatsService) GetTicketTypeStats() ([]TicketTypeStats, error) {
	var stats []TicketTypeStats

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

	return stats, nil
}

func (s *StatsService) GetConversionFunnel(eventID uint) (*ConversionFunnel, error) {
	funnel := &ConversionFunnel{}

	// 页面浏览数（简化：使用活动浏览记录，这里用票务查询代替）
	s.db.Model(&db.Ticket{}).Where("event_id = ?", eventID).Count(&funnel.PageViews)

	// 加入购物车（排队/等候名单）
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "reserved").Count(&funnel.AddToCart)

	// 预定
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "reserved").Count(&funnel.Reserved)

	// 支付
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "paid").Count(&funnel.Paid)

	// 使用
	s.db.Model(&db.Ticket{}).Where("event_id = ? AND status = ?", eventID, "used").Count(&funnel.Used)

	return funnel, nil
}
