package repository

import (
	"time"

	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type PromoCodeRepository interface {
	Create(promoCode *db.PromoCode) error
	FindByCode(code string) (*db.PromoCode, error)
	FindByEventID(eventID uint) ([]db.PromoCode, error)
	Update(promoCode *db.PromoCode) error
	Delete(id uint) error
	IncrementUsedCount(id uint) error
}

type promoCodeRepository struct {
	db *gorm.DB
}

func NewPromoCodeRepository(db *gorm.DB) PromoCodeRepository {
	return &promoCodeRepository{db: db}
}

func (r *promoCodeRepository) Create(promoCode *db.PromoCode) error {
	return r.db.Create(promoCode).Error
}

func (r *promoCodeRepository) FindByCode(code string) (*db.PromoCode, error) {
	var promoCode db.PromoCode
	if err := r.db.Where("code = ? AND is_active = true", code).First(&promoCode).Error; err != nil {
		return nil, err
	}
	return &promoCode, nil
}

func (r *promoCodeRepository) FindByEventID(eventID uint) ([]db.PromoCode, error) {
	var promoCodes []db.PromoCode
	if err := r.db.Where("event_id = ? OR event_id = 0", eventID).Find(&promoCodes).Error; err != nil {
		return nil, err
	}
	return promoCodes, nil
}

func (r *promoCodeRepository) Update(promoCode *db.PromoCode) error {
	return r.db.Save(promoCode).Error
}

func (r *promoCodeRepository) Delete(id uint) error {
	return r.db.Delete(&db.PromoCode{}, id).Error
}

func (r *promoCodeRepository) IncrementUsedCount(id uint) error {
	return r.db.Model(&db.PromoCode{}).Where("id = ?", id).Update("used_count", gorm.Expr("used_count + 1")).Error
}

// ValidatePromoCode 验证促销码是否有效
func ValidatePromoCode(promoCode *db.PromoCode, amount float64) error {
	now := time.Now()

	// 检查是否在有效期内
	if !promoCode.StartTime.IsZero() && now.Before(promoCode.StartTime) {
		return gorm.ErrRecordNotFound
	}
	if !promoCode.EndTime.IsZero() && now.After(promoCode.EndTime) {
		return gorm.ErrRecordNotFound
	}

	// 检查使用次数
	if promoCode.MaxUses > 0 && promoCode.UsedCount >= promoCode.MaxUses {
		return gorm.ErrRecordNotFound
	}

	// 检查最低消费金额
	if amount < promoCode.MinAmount {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CalculateDiscount 计算折扣金额
func CalculateDiscount(promoCode *db.PromoCode, amount float64) float64 {
	switch promoCode.DiscountType {
	case "percent":
		discount := amount * promoCode.DiscountValue / 100
		return discount
	case "fixed":
		return promoCode.DiscountValue
	default:
		return 0
	}
}
