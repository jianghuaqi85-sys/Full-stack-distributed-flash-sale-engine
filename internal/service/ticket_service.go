package service

import (
	"context"
	"fmt"

	"order-system/internal/mq"
	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/repository"
)

type TicketService struct {
	redis          *pkgredis.RedisClientWrapper
	producer       *mq.TicketProducer
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	eventRepo      repository.EventRepository
}

func NewTicketService(
	redis *pkgredis.RedisClientWrapper,
	producer *mq.TicketProducer,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	eventRepo repository.EventRepository,
) *TicketService {
	return &TicketService{
		redis:          redis,
		producer:       producer,
		ticketRepo:     ticketRepo,
		ticketTypeRepo: ticketTypeRepo,
		eventRepo:      eventRepo,
	}
}

type PurchaseTicketInput struct {
	UserID       uint
	EventID      uint
	ShowID       uint
	TicketTypeID uint
	Quantity     int
}

type TicketResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	OrderNo string `json:"order_no,omitempty"`
}

type TicketOutput struct {
	ID         uint    `json:"id"`
	OrderNo    string  `json:"order_no"`
	EventID    uint    `json:"event_id"`
	TicketName string  `json:"ticket_name"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
	QRCode     string  `json:"qr_code,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

var (
	ErrTicketSoldOut    = fmt.Errorf("已售罄")
	ErrTicketDuplicate  = fmt.Errorf("您已购买过该票种")
	ErrEventNotOnSale   = fmt.Errorf("活动未在售票")
	ErrEventNotStarted  = fmt.Errorf("活动已结束")
)

func (s *TicketService) PurchaseTicket(ctx context.Context, input *PurchaseTicketInput) (*TicketResult, error) {
	event, err := s.eventRepo.FindByID(input.EventID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("活动不存在")
	}

	if event.Status != constants.EventStatusOnSale {
		return nil, ErrEventNotOnSale
	}

	ticketType, err := s.ticketTypeRepo.FindByID(input.TicketTypeID)
	if err != nil || ticketType == nil {
		return nil, fmt.Errorf("票种不存在")
	}

	if ticketType.EventID != input.EventID {
		return nil, fmt.Errorf("票种不属于该活动")
	}

	if input.Quantity > ticketType.MaxPerUser {
		return nil, fmt.Errorf("超出每用户限购数量 %d", ticketType.MaxPerUser)
	}

	activityID := fmt.Sprintf("ticket:%d", input.EventID)

	result, err := s.redis.SeckillDeduct(ctx, activityID,
		fmt.Sprint(input.TicketTypeID), fmt.Sprint(input.UserID))
	if err != nil {
		return nil, fmt.Errorf("购票失败: %w", err)
	}

	switch result {
	case -1:
		return nil, ErrTicketSoldOut
	case -2:
		return nil, ErrTicketDuplicate
	}

	orderMsg := mq.NewTicketMessage(input.UserID, input.EventID, input.TicketTypeID, input.Quantity)
	if err := s.producer.PublishTicket(ctx, orderMsg); err != nil {
		s.redis.SeckillRollback(ctx, activityID,
			fmt.Sprint(input.TicketTypeID), fmt.Sprint(input.UserID))
		return nil, fmt.Errorf("下单失败: %w", err)
	}

	return &TicketResult{
		Status:  "queued",
		Message: "排队中，请稍候...",
	}, nil
}

func (s *TicketService) GetMyTickets(userID uint, page, limit int) ([]TicketOutput, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	tickets, total, err := s.ticketRepo.FindByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// 批量加载 TicketType，避免 N+1 查询
	ticketTypeIDs := make([]uint, 0, len(tickets))
	seen := make(map[uint]bool)
	for _, ticket := range tickets {
		if !seen[ticket.TicketTypeID] {
			ticketTypeIDs = append(ticketTypeIDs, ticket.TicketTypeID)
			seen[ticket.TicketTypeID] = true
		}
	}

	ticketTypeMap := make(map[uint]string)
	if len(ticketTypeIDs) > 0 {
		tts, _ := s.ticketTypeRepo.FindByIDs(ticketTypeIDs)
		for _, tt := range tts {
			ticketTypeMap[tt.ID] = tt.Name
		}
	}

	output := make([]TicketOutput, 0, len(tickets))
	for _, ticket := range tickets {
		output = append(output, TicketOutput{
			ID:         ticket.ID,
			OrderNo:    ticket.OrderNo,
			EventID:    ticket.EventID,
			TicketName: ticketTypeMap[ticket.TicketTypeID],
			Quantity:   ticket.Quantity,
			TotalPrice: ticket.TotalPrice,
			Status:     ticket.Status,
			QRCode:     ticket.QRCode,
			CreatedAt:  ticket.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return output, total, nil
}

func (s *TicketService) GetTicketDetail(userID, ticketID uint) (*TicketOutput, error) {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil || ticket == nil {
		return nil, fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	tt, _ := s.ticketTypeRepo.FindByID(ticket.TicketTypeID)
	ticketName := ""
	if tt != nil {
		ticketName = tt.Name
	}

	return &TicketOutput{
		ID:         ticket.ID,
		OrderNo:    ticket.OrderNo,
		EventID:    ticket.EventID,
		TicketName: ticketName,
		Quantity:   ticket.Quantity,
		TotalPrice: ticket.TotalPrice,
		Status:     ticket.Status,
		QRCode:     ticket.QRCode,
		CreatedAt:  ticket.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *TicketService) PayTicket(userID, ticketID uint) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil || ticket == nil {
		return fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if err := constants.IsValidTicketTransition(ticket.Status, constants.TicketStatusPaid); err != nil {
		return err
	}

	return s.ticketRepo.UpdateStatus(ticketID, constants.TicketStatusPaid)
}

func (s *TicketService) CancelTicket(ctx context.Context, userID, ticketID uint) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil || ticket == nil {
		return fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if err := constants.IsValidTicketTransition(ticket.Status, constants.TicketStatusCancelled); err != nil {
		return err
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, constants.TicketStatusCancelled); err != nil {
		return err
	}

	// 回滚数据库库存
	if err := s.ticketTypeRepo.UpdateStock(ticket.TicketTypeID, -ticket.Quantity); err != nil {
		return fmt.Errorf("failed to rollback stock: %w", err)
	}

	// 回滚 Redis 库存
	activityID := fmt.Sprintf("ticket:%d", ticket.EventID)
	s.redis.SeckillRollback(ctx, activityID,
		fmt.Sprint(ticket.TicketTypeID), fmt.Sprint(userID))

	return nil
}

func (s *TicketService) UseTicket(userID, ticketID uint) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil || ticket == nil {
		return fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if err := constants.IsValidTicketTransition(ticket.Status, constants.TicketStatusUsed); err != nil {
		return err
	}

	return s.ticketRepo.UpdateStatus(ticketID, constants.TicketStatusUsed)
}
