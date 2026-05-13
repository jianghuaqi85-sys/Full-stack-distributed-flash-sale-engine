package repository

import (
	"fmt"

	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type TicketTypeRepository interface {
	Create(tt *db.TicketType) error
	FindByID(id uint) (*db.TicketType, error)
	FindByIDs(ids []uint) ([]db.TicketType, error)
	FindByEventID(eventID uint) ([]db.TicketType, error)
	Update(tt *db.TicketType) error
	Delete(id uint) error
	UpdateStock(id uint, quantity int) error
	AtomicDeductStock(id uint, quantity int) error
}

type ticketTypeRepository struct {
	db *gorm.DB
}

func NewTicketTypeRepository(db *gorm.DB) TicketTypeRepository {
	return &ticketTypeRepository{db: db}
}

func (r *ticketTypeRepository) Create(tt *db.TicketType) error {
	return r.db.Create(tt).Error
}

func (r *ticketTypeRepository) FindByID(id uint) (*db.TicketType, error) {
	var tt db.TicketType
	if err := r.db.First(&tt, id).Error; err != nil {
		return nil, err
	}
	return &tt, nil
}

func (r *ticketTypeRepository) FindByIDs(ids []uint) ([]db.TicketType, error) {
	var ticketTypes []db.TicketType
	if err := r.db.Where("id IN ?", ids).Find(&ticketTypes).Error; err != nil {
		return nil, err
	}
	return ticketTypes, nil
}

func (r *ticketTypeRepository) FindByEventID(eventID uint) ([]db.TicketType, error) {
	var ticketTypes []db.TicketType
	if err := r.db.Where("event_id = ?", eventID).Order("sort_order ASC").Find(&ticketTypes).Error; err != nil {
		return nil, err
	}
	return ticketTypes, nil
}

func (r *ticketTypeRepository) Update(tt *db.TicketType) error {
	return r.db.Save(tt).Error
}

func (r *ticketTypeRepository) Delete(id uint) error {
	return r.db.Delete(&db.TicketType{}, id).Error
}

func (r *ticketTypeRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&db.TicketType{}).Where("id = ?", id).Update("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *ticketTypeRepository) AtomicDeductStock(id uint, quantity int) error {
	result := r.db.Model(&db.TicketType{}).
		Where("id = ? AND stock >= ?", id, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.RowsAffected == 0 {
		return fmt.Errorf("库存不足")
	}
	return result.Error
}
