package repository

import (
	"time"

	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type TicketRepository interface {
	Create(ticket *db.Ticket) error
	FindByID(id uint) (*db.Ticket, error)
	FindByUserID(userID uint, page, limit int) ([]db.Ticket, int64, error)
	FindByEventID(eventID uint, page, limit int) ([]db.Ticket, int64, error)
	FindByStatus(status string, page, limit int) ([]db.Ticket, int64, error)
	FindExpiredReserved(olderThan time.Time, limit int) ([]db.Ticket, error)
	UpdateStatus(id uint, status string) error
	UpdateOwner(id uint, newUserID uint) error
	UpdateOrderNo(id uint, orderNo string) error
	CountByUserAndTicketType(userID, ticketTypeID uint) (int64, error)
	ExistsByOrderNo(orderNo string) (bool, error)
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *db.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepository) FindByID(id uint) (*db.Ticket, error) {
	var ticket db.Ticket
	if err := r.db.First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) FindByUserID(userID uint, page, limit int) ([]db.Ticket, int64, error) {
	var tickets []db.Ticket
	var total int64

	query := r.db.Model(&db.Ticket{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *ticketRepository) FindByEventID(eventID uint, page, limit int) ([]db.Ticket, int64, error) {
	var tickets []db.Ticket
	var total int64

	query := r.db.Model(&db.Ticket{}).Where("event_id = ?", eventID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *ticketRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&db.Ticket{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ticketRepository) UpdateOrderNo(id uint, orderNo string) error {
	return r.db.Model(&db.Ticket{}).Where("id = ?", id).Update("order_no", orderNo).Error
}

func (r *ticketRepository) CountByUserAndTicketType(userID, ticketTypeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&db.Ticket{}).
		Where("user_id = ? AND ticket_type_id = ? AND status != ?", userID, ticketTypeID, "cancelled").
		Count(&count).Error
	return count, err
}

func (r *ticketRepository) ExistsByOrderNo(orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&db.Ticket{}).Where("order_no = ?", orderNo).Count(&count).Error
	return count > 0, err
}

func (r *ticketRepository) FindByStatus(status string, page, limit int) ([]db.Ticket, int64, error) {
	var tickets []db.Ticket
	var total int64

	query := r.db.Model(&db.Ticket{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at ASC").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *ticketRepository) FindExpiredReserved(olderThan time.Time, limit int) ([]db.Ticket, error) {
	var tickets []db.Ticket
	err := r.db.Where("status = ? AND created_at < ?", "reserved", olderThan).
		Order("created_at ASC").Limit(limit).Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) UpdateOwner(id uint, newUserID uint) error {
	return r.db.Model(&db.Ticket{}).Where("id = ?", id).
		Updates(map[string]interface{}{"user_id": newUserID, "transfer_status": "approved", "transferred_to": newUserID}).Error
}
