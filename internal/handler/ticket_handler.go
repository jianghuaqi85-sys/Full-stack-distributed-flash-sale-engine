package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/pkg/db"
	"order-system/internal/service"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) PurchaseTicket(c *gin.Context) {
	var req struct {
		EventID      uint `json:"event_id" binding:"required"`
		ShowID       uint `json:"show_id"`
		TicketTypeID uint `json:"ticket_type_id" binding:"required"`
		Quantity     int  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	result, err := h.ticketService.PurchaseTicket(c.Request.Context(), &service.PurchaseTicketInput{
		UserID:       userModel.ID,
		EventID:      req.EventID,
		ShowID:       req.ShowID,
		TicketTypeID: req.TicketTypeID,
		Quantity:     req.Quantity,
	})

	if err != nil {
		switch err {
		case service.ErrTicketSoldOut:
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		case service.ErrTicketDuplicate:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case service.ErrEventNotOnSale:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  result.Status,
		"message": result.Message,
	})
}

func (h *TicketHandler) GetMyTickets(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	tickets, total, err := h.ticketService.GetMyTickets(userModel.ID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tickets,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *TicketHandler) GetTicketDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的票务ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	ticket, err := h.ticketService.GetTicketDetail(userModel.ID, uint(id))
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) PayTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的票务ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	if err := h.ticketService.PayTicket(userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "支付成功"})
}

func (h *TicketHandler) CancelTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的票务ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	if err := h.ticketService.CancelTicket(c.Request.Context(), userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消成功"})
}

func (h *TicketHandler) UseTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的票务ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return
	}

	if err := h.ticketService.UseTicket(userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "使用成功"})
}
