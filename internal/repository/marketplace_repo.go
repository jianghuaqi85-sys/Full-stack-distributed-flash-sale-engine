package repository

import (
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type MarketplaceRepository interface {
	Create(listing *db.MarketplaceListing) error
	FindByID(id uint) (*db.MarketplaceListing, error)
	FindByTicketID(ticketID uint) (*db.MarketplaceListing, error)
	FindActiveByEventID(eventID uint, page, limit int) ([]db.MarketplaceListing, int64, error)
	FindActiveListings(page, limit int) ([]db.MarketplaceListing, int64, error)
	FindBySellerID(userID uint) ([]db.MarketplaceListing, error)
	FindByBuyerID(userID uint) ([]db.MarketplaceListing, error)
	FindBySellerIDPaginated(userID uint, page, limit int) ([]db.MarketplaceListing, int64, error)
	FindByBuyerIDPaginated(userID uint, page, limit int) ([]db.MarketplaceListing, int64, error)
	Update(listing *db.MarketplaceListing) error
}

type marketplaceRepository struct {
	db *gorm.DB
}

func NewMarketplaceRepository(db *gorm.DB) MarketplaceRepository {
	return &marketplaceRepository{db: db}
}

func (r *marketplaceRepository) Create(listing *db.MarketplaceListing) error {
	return r.db.Create(listing).Error
}

func (r *marketplaceRepository) FindByID(id uint) (*db.MarketplaceListing, error) {
	var listing db.MarketplaceListing
	if err := r.db.First(&listing, id).Error; err != nil {
		return nil, err
	}
	return &listing, nil
}

func (r *marketplaceRepository) FindByTicketID(ticketID uint) (*db.MarketplaceListing, error) {
	var listing db.MarketplaceListing
	if err := r.db.Where("ticket_id = ? AND status = ?", ticketID, "active").First(&listing).Error; err != nil {
		return nil, err
	}
	return &listing, nil
}

func (r *marketplaceRepository) FindActiveByEventID(eventID uint, page, limit int) ([]db.MarketplaceListing, int64, error) {
	var listings []db.MarketplaceListing
	var total int64

	query := r.db.Model(&db.MarketplaceListing{}).
		Joins("JOIN tickets ON tickets.id = marketplace_listings.ticket_id").
		Where("tickets.event_id = ? AND marketplace_listings.status = ?", eventID, "active")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("marketplace_listings.created_at DESC").Find(&listings).Error; err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *marketplaceRepository) FindActiveListings(page, limit int) ([]db.MarketplaceListing, int64, error) {
	var listings []db.MarketplaceListing
	var total int64

	query := r.db.Model(&db.MarketplaceListing{}).Where("status = ?", "active")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&listings).Error; err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *marketplaceRepository) FindBySellerID(userID uint) ([]db.MarketplaceListing, error) {
	var listings []db.MarketplaceListing
	if err := r.db.Where("seller_id = ?", userID).Order("created_at DESC").Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

func (r *marketplaceRepository) FindByBuyerID(userID uint) ([]db.MarketplaceListing, error) {
	var listings []db.MarketplaceListing
	if err := r.db.Where("buyer_id = ?", userID).Order("created_at DESC").Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

func (r *marketplaceRepository) Update(listing *db.MarketplaceListing) error {
	return r.db.Save(listing).Error
}

func (r *marketplaceRepository) FindBySellerIDPaginated(userID uint, page, limit int) ([]db.MarketplaceListing, int64, error) {
	var listings []db.MarketplaceListing
	var total int64

	query := r.db.Model(&db.MarketplaceListing{}).Where("seller_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&listings).Error; err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *marketplaceRepository) FindByBuyerIDPaginated(userID uint, page, limit int) ([]db.MarketplaceListing, int64, error) {
	var listings []db.MarketplaceListing
	var total int64

	query := r.db.Model(&db.MarketplaceListing{}).Where("buyer_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&listings).Error; err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}
