package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
		Quantity     int  `json:"quantity" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
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
	page, limit := parsePageLimit(c, 1, 10, 100)

	userModel, ok := getUser(c)
	if !ok {
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

	userModel, ok := getUser(c)
	if !ok {
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

	userModel, ok := getUser(c)
	if !ok {
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

	userModel, ok := getUser(c)
	if !ok {
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

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	if err := h.ticketService.UseTicket(userModel.ID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "使用成功"})
}
