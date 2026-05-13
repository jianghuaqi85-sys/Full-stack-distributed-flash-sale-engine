package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type MarketplaceHandler struct {
	marketplaceService *service.MarketplaceService
}

func NewMarketplaceHandler(marketplaceService *service.MarketplaceService) *MarketplaceHandler {
	return &MarketplaceHandler{marketplaceService: marketplaceService}
}

func (h *MarketplaceHandler) CreateListing(c *gin.Context) {
	var req struct {
		TicketID    uint    `json:"ticket_id" binding:"required"`
		Price       float64 `json:"price" binding:"required,min=0.01"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	listing, err := h.marketplaceService.CreateListing(userModel.ID, service.CreateListingInput{
		TicketID:    req.TicketID,
		Price:       req.Price,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      listing.ID,
		"status":  listing.Status,
		"message": "上架成功",
	})
}

func (h *MarketplaceHandler) BuyListing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listing id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	if err := h.marketplaceService.BuyListing(userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "购买成功"})
}

func (h *MarketplaceHandler) CancelListing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listing id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	if err := h.marketplaceService.CancelListing(userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "下架成功"})
}

func (h *MarketplaceHandler) GetListing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listing id"})
		return
	}

	listing, err := h.marketplaceService.GetListing(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, listing)
}

func (h *MarketplaceHandler) ListActive(c *gin.Context) {
	page, limit := parsePageLimit(c, 1, 20, 100)

	listings, total, err := h.marketplaceService.ListActiveListings(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  listings,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MarketplaceHandler) ListByEvent(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	page, limit := parsePageLimit(c, 1, 20, 100)

	listings, total, err := h.marketplaceService.ListByEvent(uint(eventID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  listings,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MarketplaceHandler) ListMyListings(c *gin.Context) {
	userModel, ok := getUser(c)
	if !ok {
		return
	}

	listings, err := h.marketplaceService.ListMyListings(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": listings})
}

func (h *MarketplaceHandler) ListMyPurchases(c *gin.Context) {
	userModel, ok := getUser(c)
	if !ok {
		return
	}

	listings, err := h.marketplaceService.ListMyPurchases(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": listings})
}
