package service

import (
	"context"
	"fmt"
	"time"

	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/db"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/repository"
)

type EventService struct {
	eventRepo      repository.EventRepository
	ticketTypeRepo repository.TicketTypeRepository
	redis          *pkgredis.RedisClientWrapper
}

func NewEventService(eventRepo repository.EventRepository, ticketTypeRepo repository.TicketTypeRepository, redis *pkgredis.RedisClientWrapper) *EventService {
	return &EventService{
		eventRepo:      eventRepo,
		ticketTypeRepo: ticketTypeRepo,
		redis:          redis,
	}
}

type CreateEventInput struct {
	Title       string
	Description string
	Location    string
	CoverImage  string
	StartTime   time.Time
	EndTime     time.Time
}

type UpdateEventInput struct {
	Title       string
	Description string
	Location    string
	CoverImage  string
	StartTime   time.Time
	EndTime     time.Time
}

type CreateTicketTypeInput struct {
	Name       string
	Price      float64
	Stock      int
	MaxPerUser int
	SortOrder  int
}

type EventOutput struct {
	ID          uint               `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Location    string             `json:"location"`
	CoverImage  string             `json:"cover_image"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     time.Time          `json:"end_time"`
	Status      string             `json:"status"`
	TotalStock  int                `json:"total_stock"`
	TicketTypes []TicketTypeOutput `json:"ticket_types,omitempty"`
}

type TicketTypeOutput struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Stock      int     `json:"stock"`
	MaxPerUser int     `json:"max_per_user"`
	SortOrder  int     `json:"sort_order"`
}

func (s *EventService) CreateEvent(input CreateEventInput) (*db.Event, error) {
	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	event := &db.Event{
		Title:       input.Title,
		Description: input.Description,
		Location:    input.Location,
		CoverImage:  input.CoverImage,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      constants.EventStatusDraft,
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

func (s *EventService) UpdateEvent(id uint, input UpdateEventInput) (*db.Event, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}

	if event.Status != constants.EventStatusDraft {
		return nil, fmt.Errorf("can only edit draft events")
	}

	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Location = input.Location
	event.CoverImage = input.CoverImage
	event.StartTime = input.StartTime
	event.EndTime = input.EndTime

	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

func (s *EventService) GetEvent(id uint) (*EventOutput, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEventID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket types: %w", err)
	}

	output := &EventOutput{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		CoverImage:  event.CoverImage,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
		Status:      event.Status,
		TotalStock:  event.TotalStock,
	}

	for _, tt := range ticketTypes {
		output.TicketTypes = append(output.TicketTypes, TicketTypeOutput{
			ID:         tt.ID,
			Name:       tt.Name,
			Price:      tt.Price,
			Stock:      tt.Stock,
			MaxPerUser: tt.MaxPerUser,
			SortOrder:  tt.SortOrder,
		})
	}

	return output, nil
}

func (s *EventService) ListEvents(page, limit int) ([]EventOutput, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	events, total, err := s.eventRepo.FindAll(page, limit)
	if err != nil {
		return nil, 0, err
	}

	output := make([]EventOutput, 0, len(events))
	for _, event := range events {
		output = append(output, EventOutput{
			ID:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			Location:    event.Location,
			CoverImage:  event.CoverImage,
			StartTime:   event.StartTime,
			EndTime:     event.EndTime,
			Status:      event.Status,
			TotalStock:  event.TotalStock,
		})
	}

	return output, total, nil
}

func (s *EventService) PublishEvent(id uint) error {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}

	if err := constants.IsValidEventTransition(event.Status, constants.EventStatusOnSale); err != nil {
		return err
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEventID(id)
	if err != nil {
		return fmt.Errorf("failed to get ticket types: %w", err)
	}

	if len(ticketTypes) == 0 {
		return fmt.Errorf("event must have at least one ticket type")
	}

	totalStock := 0
	for _, tt := range ticketTypes {
		totalStock += tt.Stock
	}

	if err := s.eventRepo.UpdateStatus(id, constants.EventStatusOnSale); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	if err := s.eventRepo.UpdateTotalStock(id, totalStock); err != nil {
		return fmt.Errorf("failed to update total stock: %w", err)
	}

	// 初始化 Redis 库存
	ctx := context.Background()
	activityID := fmt.Sprintf("ticket:%d", id)
	for _, tt := range ticketTypes {
		if err := s.redis.InitSeckillStock(ctx, activityID, fmt.Sprint(tt.ID), tt.Stock); err != nil {
			return fmt.Errorf("failed to init redis stock for ticket type %d: %w", tt.ID, err)
		}
	}

	return nil
}

func (s *EventService) UnpublishEvent(id uint) error {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}

	if err := constants.IsValidEventTransition(event.Status, constants.EventStatusOffSale); err != nil {
		return err
	}

	return s.eventRepo.UpdateStatus(id, constants.EventStatusOffSale)
}

func (s *EventService) EndEvent(id uint) error {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}

	if err := constants.IsValidEventTransition(event.Status, constants.EventStatusEnded); err != nil {
		return err
	}

	return s.eventRepo.UpdateStatus(id, constants.EventStatusEnded)
}

func (s *EventService) CreateTicketType(eventID uint, input CreateTicketTypeInput) (*db.TicketType, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}

	if event.Status != constants.EventStatusDraft {
		return nil, fmt.Errorf("can only add ticket types to draft events")
	}

	if input.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	if input.Stock < 0 {
		return nil, fmt.Errorf("stock must be non-negative")
	}

	tt := &db.TicketType{
		EventID:    eventID,
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		MaxPerUser: input.MaxPerUser,
		SortOrder:  input.SortOrder,
	}

	if tt.MaxPerUser <= 0 {
		tt.MaxPerUser = 1
	}

	if err := s.ticketTypeRepo.Create(tt); err != nil {
		return nil, fmt.Errorf("failed to create ticket type: %w", err)
	}

	return tt, nil
}

func (s *EventService) UpdateTicketType(id uint, input CreateTicketTypeInput) (*db.TicketType, error) {
	tt, err := s.ticketTypeRepo.FindByID(id)
	if err != nil || tt == nil {
		return nil, fmt.Errorf("ticket type not found")
	}

	event, err := s.eventRepo.FindByID(tt.EventID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}

	if event.Status != constants.EventStatusDraft {
		return nil, fmt.Errorf("can only edit ticket types of draft events")
	}

	tt.Name = input.Name
	tt.Price = input.Price
	tt.Stock = input.Stock
	tt.MaxPerUser = input.MaxPerUser
	tt.SortOrder = input.SortOrder

	if tt.MaxPerUser <= 0 {
		tt.MaxPerUser = 1
	}

	if err := s.ticketTypeRepo.Update(tt); err != nil {
		return nil, fmt.Errorf("failed to update ticket type: %w", err)
	}

	return tt, nil
}

func (s *EventService) DeleteTicketType(id uint) error {
	tt, err := s.ticketTypeRepo.FindByID(id)
	if err != nil || tt == nil {
		return fmt.Errorf("ticket type not found")
	}

	event, err := s.eventRepo.FindByID(tt.EventID)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}

	if event.Status != constants.EventStatusDraft {
		return fmt.Errorf("can only delete ticket types of draft events")
	}

	return s.ticketTypeRepo.Delete(id)
}

func (s *EventService) GetEventStock(id uint) (map[string]int, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEventID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket types: %w", err)
	}

	stock := make(map[string]int)
	for _, tt := range ticketTypes {
		stock[fmt.Sprint(tt.ID)] = tt.Stock
	}

	return stock, nil
}
