package service

import (
	"fmt"
	"time"

	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type PromoCodeService struct {
	promoCodeRepo repository.PromoCodeRepository
}

func NewPromoCodeService(promoCodeRepo repository.PromoCodeRepository) *PromoCodeService {
	return &PromoCodeService{promoCodeRepo: promoCodeRepo}
}

type CreatePromoCodeInput struct {
	Code          string
	EventID       uint
	DiscountType  string
	DiscountValue float64
	MinAmount     float64
	MaxUses       int
	StartTime     time.Time
	EndTime       time.Time
}

type PromoCodeOutput struct {
	ID            uint      `json:"id"`
	Code          string    `json:"code"`
	EventID       uint      `json:"event_id"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinAmount     float64   `json:"min_amount"`
	MaxUses       int       `json:"max_uses"`
	UsedCount     int       `json:"used_count"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsActive      bool      `json:"is_active"`
}

func (s *PromoCodeService) CreatePromoCode(input CreatePromoCodeInput) (*db.PromoCode, error) {
	// 检查促销码是否已存在
	existing, _ := s.promoCodeRepo.FindByCode(input.Code)
	if existing != nil {
		return nil, fmt.Errorf("促销码已存在")
	}

	promoCode := &db.PromoCode{
		Code:          input.Code,
		EventID:       input.EventID,
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
		MinAmount:     input.MinAmount,
		MaxUses:       input.MaxUses,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		IsActive:      true,
	}

	if err := s.promoCodeRepo.Create(promoCode); err != nil {
		return nil, fmt.Errorf("创建促销码失败: %w", err)
	}

	return promoCode, nil
}

func (s *PromoCodeService) ValidatePromoCode(code string, amount float64) (*db.PromoCode, error) {
	promoCode, err := s.promoCodeRepo.FindByCode(code)
	if err != nil {
		return nil, fmt.Errorf("促销码不存在或已失效")
	}

	if err := repository.ValidatePromoCode(promoCode, amount); err != nil {
		return nil, fmt.Errorf("促销码不满足使用条件")
	}

	return promoCode, nil
}

func (s *PromoCodeService) CalculateDiscount(promoCode *db.PromoCode, amount float64) float64 {
	return repository.CalculateDiscount(promoCode, amount)
}

func (s *PromoCodeService) UsePromoCode(id uint) error {
	return s.promoCodeRepo.IncrementUsedCount(id)
}

func (s *PromoCodeService) GetPromoCodesByEvent(eventID uint) ([]PromoCodeOutput, error) {
	promoCodes, err := s.promoCodeRepo.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}

	output := make([]PromoCodeOutput, 0, len(promoCodes))
	for _, pc := range promoCodes {
		output = append(output, PromoCodeOutput{
			ID:            pc.ID,
			Code:          pc.Code,
			EventID:       pc.EventID,
			DiscountType:  pc.DiscountType,
			DiscountValue: pc.DiscountValue,
			MinAmount:     pc.MinAmount,
			MaxUses:       pc.MaxUses,
			UsedCount:     pc.UsedCount,
			StartTime:     pc.StartTime,
			EndTime:       pc.EndTime,
			IsActive:      pc.IsActive,
		})
	}

	return output, nil
}

func (s *PromoCodeService) DeletePromoCode(id uint) error {
	return s.promoCodeRepo.Delete(id)
}
