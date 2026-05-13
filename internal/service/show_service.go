package service

import (
	"fmt"
	"time"

	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type ShowService struct {
	showRepo      repository.ShowRepository
	eventRepo     repository.EventRepository
	ticketTypeRepo repository.TicketTypeRepository
}

func NewShowService(showRepo repository.ShowRepository, eventRepo repository.EventRepository, ticketTypeRepo repository.TicketTypeRepository) *ShowService {
	return &ShowService{
		showRepo:       showRepo,
		eventRepo:      eventRepo,
		ticketTypeRepo: ticketTypeRepo,
	}
}

type CreateShowInput struct {
	EventID   uint
	Name      string
	ShowTime  time.Time
	EndTime   time.Time
	Stock     int
	SortOrder int
}

type UpdateShowInput struct {
	Name      string
	ShowTime  time.Time
	EndTime   time.Time
	Stock     int
	SortOrder int
}

type ShowOutput struct {
	ID        uint      `json:"id"`
	EventID   uint      `json:"event_id"`
	Name      string    `json:"name"`
	ShowTime  time.Time `json:"show_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	Stock     int       `json:"stock"`
	SoldCount int       `json:"sold_count"`
	SortOrder int       `json:"sort_order"`
}

func (s *ShowService) CreateShow(input CreateShowInput) (*db.Show, error) {
	event, err := s.eventRepo.FindByID(input.EventID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("活动不存在")
	}

	if input.EndTime.Before(input.ShowTime) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	show := &db.Show{
		EventID:   input.EventID,
		Name:      input.Name,
		ShowTime:  input.ShowTime,
		EndTime:   input.EndTime,
		Status:    constants.EventStatusDraft,
		Stock:     input.Stock,
		SortOrder: input.SortOrder,
	}

	if err := s.showRepo.Create(show); err != nil {
		return nil, fmt.Errorf("创建场次失败: %w", err)
	}

	return show, nil
}

func (s *ShowService) UpdateShow(id uint, input UpdateShowInput) (*db.Show, error) {
	show, err := s.showRepo.FindByID(id)
	if err != nil || show == nil {
		return nil, fmt.Errorf("场次不存在")
	}

	if show.Status != constants.EventStatusDraft {
		return nil, fmt.Errorf("只能编辑草稿状态的场次")
	}

	if input.EndTime.Before(input.ShowTime) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	show.Name = input.Name
	show.ShowTime = input.ShowTime
	show.EndTime = input.EndTime
	show.Stock = input.Stock
	show.SortOrder = input.SortOrder

	if err := s.showRepo.Update(show); err != nil {
		return nil, fmt.Errorf("更新场次失败: %w", err)
	}

	return show, nil
}

func (s *ShowService) DeleteShow(id uint) error {
	show, err := s.showRepo.FindByID(id)
	if err != nil || show == nil {
		return fmt.Errorf("场次不存在")
	}

	if show.Status != constants.EventStatusDraft {
		return fmt.Errorf("只能删除草稿状态的场次")
	}

	return s.showRepo.Delete(id)
}

func (s *ShowService) PublishShow(id uint) error {
	show, err := s.showRepo.FindByID(id)
	if err != nil || show == nil {
		return fmt.Errorf("场次不存在")
	}

	if err := constants.IsValidEventTransition(show.Status, constants.EventStatusOnSale); err != nil {
		return err
	}

	return s.showRepo.UpdateStatus(id, constants.EventStatusOnSale)
}

func (s *ShowService) UnpublishShow(id uint) error {
	show, err := s.showRepo.FindByID(id)
	if err != nil || show == nil {
		return fmt.Errorf("场次不存在")
	}

	if err := constants.IsValidEventTransition(show.Status, constants.EventStatusOffSale); err != nil {
		return err
	}

	return s.showRepo.UpdateStatus(id, constants.EventStatusOffSale)
}

func (s *ShowService) GetShow(id uint) (*ShowOutput, error) {
	show, err := s.showRepo.FindByID(id)
	if err != nil || show == nil {
		return nil, fmt.Errorf("场次不存在")
	}

	return &ShowOutput{
		ID:        show.ID,
		EventID:   show.EventID,
		Name:      show.Name,
		ShowTime:  show.ShowTime,
		EndTime:   show.EndTime,
		Status:    show.Status,
		Stock:     show.Stock,
		SoldCount: show.SoldCount,
		SortOrder: show.SortOrder,
	}, nil
}

func (s *ShowService) ListShowsByEvent(eventID uint) ([]ShowOutput, error) {
	shows, err := s.showRepo.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}

	output := make([]ShowOutput, 0, len(shows))
	for _, show := range shows {
		output = append(output, ShowOutput{
			ID:        show.ID,
			EventID:   show.EventID,
			Name:      show.Name,
			ShowTime:  show.ShowTime,
			EndTime:   show.EndTime,
			Status:    show.Status,
			Stock:     show.Stock,
			SoldCount: show.SoldCount,
			SortOrder: show.SortOrder,
		})
	}

	return output, nil
}

func (s *ShowService) DeductStock(id uint, quantity int) error {
	return s.showRepo.UpdateStock(id, -quantity)
}

func (s *ShowService) RollbackStock(id uint, quantity int) error {
	return s.showRepo.UpdateStock(id, quantity)
}
