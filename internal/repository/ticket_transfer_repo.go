package repository

import (
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type TicketTransferRepository interface {
	Create(transfer *db.TicketTransfer) error
	FindByID(id uint) (*db.TicketTransfer, error)
	FindByTicketID(ticketID uint) (*db.TicketTransfer, error)
	FindByToUserID(userID uint) ([]db.TicketTransfer, error)
	FindByUserID(userID uint) ([]db.TicketTransfer, error)
	FindPending() ([]db.TicketTransfer, error)
	Update(transfer *db.TicketTransfer) error
}

type ticketTransferRepository struct {
	db *gorm.DB
}

func NewTicketTransferRepository(db *gorm.DB) TicketTransferRepository {
	return &ticketTransferRepository{db: db}
}

func (r *ticketTransferRepository) Create(transfer *db.TicketTransfer) error {
	return r.db.Create(transfer).Error
}

func (r *ticketTransferRepository) FindByID(id uint) (*db.TicketTransfer, error) {
	var transfer db.TicketTransfer
	if err := r.db.First(&transfer, id).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *ticketTransferRepository) FindByTicketID(ticketID uint) (*db.TicketTransfer, error) {
	var transfer db.TicketTransfer
	if err := r.db.Where("ticket_id = ? AND status = ?", ticketID, "pending").First(&transfer).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *ticketTransferRepository) FindByToUserID(userID uint) ([]db.TicketTransfer, error) {
	var transfers []db.TicketTransfer
	if err := r.db.Where("to_user_id = ?", userID).Find(&transfers).Error; err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *ticketTransferRepository) FindByUserID(userID uint) ([]db.TicketTransfer, error) {
	var transfers []db.TicketTransfer
	if err := r.db.Where("from_user_id = ? OR to_user_id = ?", userID, userID).Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *ticketTransferRepository) FindPending() ([]db.TicketTransfer, error) {
	var transfers []db.TicketTransfer
	if err := r.db.Where("status = ?", "pending").Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *ticketTransferRepository) Update(transfer *db.TicketTransfer) error {
	return r.db.Save(transfer).Error
}
