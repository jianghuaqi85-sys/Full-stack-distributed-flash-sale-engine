package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type TicketTransferHandler struct {
	transferService *service.TicketTransferService
}

func NewTicketTransferHandler(transferService *service.TicketTransferService) *TicketTransferHandler {
	return &TicketTransferHandler{transferService: transferService}
}

func (h *TicketTransferHandler) RequestTransfer(c *gin.Context) {
	var req struct {
		TicketID uint   `json:"ticket_id" binding:"required"`
		ToUserID uint   `json:"to_user_id" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	transfer, err := h.transferService.RequestTransfer(userModel.ID, service.RequestTransferInput{
		TicketID: req.TicketID,
		ToUserID: req.ToUserID,
		Reason:   req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      transfer.ID,
		"status":  transfer.Status,
		"message": "转让请求已提交，等待管理员审核",
	})
}

func (h *TicketTransferHandler) DirectGift(c *gin.Context) {
	var req struct {
		TicketID uint   `json:"ticket_id" binding:"required"`
		ToUserID uint   `json:"to_user_id" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	transfer, err := h.transferService.DirectGift(userModel.ID, service.RequestTransferInput{
		TicketID: req.TicketID,
		ToUserID: req.ToUserID,
		Reason:   req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      transfer.ID,
		"status":  transfer.Status,
		"message": "转赠成功",
	})
}

func (h *TicketTransferHandler) GetTransferHistory(c *gin.Context) {
	userModel, ok := getUser(c)
	if !ok {
		return
	}

	transfers, err := h.transferService.GetTransferHistory(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transfers})
}

func (h *TicketTransferHandler) ApproveTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	if err := h.transferService.ApproveTransfer(uint(id), userModel.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "转让已批准"})
}

func (h *TicketTransferHandler) RejectTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	if err := h.transferService.RejectTransfer(uint(id), userModel.ID, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "转让已拒绝"})
}

func (h *TicketTransferHandler) GetPendingTransfers(c *gin.Context) {
	transfers, err := h.transferService.GetPendingTransfers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transfers})
}
