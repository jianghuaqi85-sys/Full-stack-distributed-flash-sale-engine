package repository

import (
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type EventRepository interface {
	Create(event *db.Event) error
	FindByID(id uint) (*db.Event, error)
	FindAll(page, limit int) ([]db.Event, int64, error)
	FindByStatus(status string, page, limit int) ([]db.Event, int64, error)
	Update(event *db.Event) error
	UpdateStatus(id uint, status string) error
	UpdateTotalStock(id uint, totalStock int) error
	Delete(id uint) error
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *db.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindByID(id uint) (*db.Event, error) {
	var event db.Event
	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindAll(page, limit int) ([]db.Event, int64, error) {
	var events []db.Event
	var total int64

	if err := r.db.Model(&db.Event{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.Offset(offset).Limit(limit).Order("start_time DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *eventRepository) FindByStatus(status string, page, limit int) ([]db.Event, int64, error) {
	var events []db.Event
	var total int64

	query := r.db.Model(&db.Event{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("start_time DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *eventRepository) Update(event *db.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&db.Event{}).Where("id = ?", id).Update("status", status).Error
}

func (r *eventRepository) UpdateTotalStock(id uint, totalStock int) error {
	return r.db.Model(&db.Event{}).Where("id = ?", id).Update("total_stock", totalStock).Error
}

func (r *eventRepository) Delete(id uint) error {
	return r.db.Delete(&db.Event{}, id).Error
}
