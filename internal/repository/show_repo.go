package repository

import (
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type ShowRepository interface {
	Create(show *db.Show) error
	FindByID(id uint) (*db.Show, error)
	FindByEventID(eventID uint) ([]db.Show, error)
	Update(show *db.Show) error
	UpdateStatus(id uint, status string) error
	UpdateStock(id uint, delta int) error
	Delete(id uint) error
}

type showRepository struct {
	db *gorm.DB
}

func NewShowRepository(db *gorm.DB) ShowRepository {
	return &showRepository{db: db}
}

func (r *showRepository) Create(show *db.Show) error {
	return r.db.Create(show).Error
}

func (r *showRepository) FindByID(id uint) (*db.Show, error) {
	var show db.Show
	if err := r.db.First(&show, id).Error; err != nil {
		return nil, err
	}
	return &show, nil
}

func (r *showRepository) FindByEventID(eventID uint) ([]db.Show, error) {
	var shows []db.Show
	if err := r.db.Where("event_id = ?", eventID).Order("sort_order ASC, show_time ASC").Find(&shows).Error; err != nil {
		return nil, err
	}
	return shows, nil
}

func (r *showRepository) Update(show *db.Show) error {
	return r.db.Save(show).Error
}

func (r *showRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&db.Show{}).Where("id = ?", id).Update("status", status).Error
}

func (r *showRepository) UpdateStock(id uint, delta int) error {
	return r.db.Model(&db.Show{}).Where("id = ? AND stock + ? >= 0", id, delta).
		Updates(map[string]interface{}{
			"stock": gorm.Expr("stock + ?", delta),
		}).Error
}

func (r *showRepository) Delete(id uint) error {
	return r.db.Delete(&db.Show{}, id).Error
}
