package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type PromoCodeHandler struct {
	promoCodeService *service.PromoCodeService
}

func NewPromoCodeHandler(promoCodeService *service.PromoCodeService) *PromoCodeHandler {
	return &PromoCodeHandler{promoCodeService: promoCodeService}
}

func (h *PromoCodeHandler) CreatePromoCode(c *gin.Context) {
	var req struct {
		Code          string  `json:"code" binding:"required"`
		EventID       uint    `json:"event_id"`
		DiscountType  string  `json:"discount_type" binding:"required,oneof=percent fixed"`
		DiscountValue float64 `json:"discount_value" binding:"required,min=0"`
		MinAmount     float64 `json:"min_amount"`
		MaxUses       int     `json:"max_uses"`
		StartTime     string  `json:"start_time"`
		EndTime       string  `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime, endTime time.Time
	var err error

	if req.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
			return
		}
	}

	if req.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
			return
		}
	}

	promoCode, err := h.promoCodeService.CreatePromoCode(service.CreatePromoCodeInput{
		Code:          req.Code,
		EventID:       req.EventID,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MinAmount:     req.MinAmount,
		MaxUses:       req.MaxUses,
		StartTime:     startTime,
		EndTime:       endTime,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             promoCode.ID,
		"code":           promoCode.Code,
		"discount_type":  promoCode.DiscountType,
		"discount_value": promoCode.DiscountValue,
	})
}

func (h *PromoCodeHandler) ValidatePromoCode(c *gin.Context) {
	var req struct {
		Code   string  `json:"code" binding:"required"`
		Amount float64 `json:"amount" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promoCode, err := h.promoCodeService.ValidatePromoCode(req.Code, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	discount := h.promoCodeService.CalculateDiscount(promoCode, req.Amount)
	finalAmount := req.Amount - discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"code":           promoCode.Code,
		"discount_type":  promoCode.DiscountType,
		"discount_value": promoCode.DiscountValue,
		"discount":       discount,
		"final_amount":   finalAmount,
	})
}

func (h *PromoCodeHandler) GetPromoCodes(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	promoCodes, err := h.promoCodeService.GetPromoCodesByEvent(uint(eventID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": promoCodes})
}

func (h *PromoCodeHandler) DeletePromoCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promo code id"})
		return
	}

	if err := h.promoCodeService.DeletePromoCode(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
