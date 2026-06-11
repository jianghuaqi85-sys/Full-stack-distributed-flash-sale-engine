package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/queue"
)

type WaitlistHandler struct {
	waitlistManager *queue.WaitlistManager
}

func NewWaitlistHandler(waitlistManager *queue.WaitlistManager) *WaitlistHandler {
	return &WaitlistHandler{waitlistManager: waitlistManager}
}

func (h *WaitlistHandler) JoinWaitlist(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	_, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	userID := strconv.FormatUint(uint64(userModel.ID), 10)

	entry, err := h.waitlistManager.JoinWaitlist(c.Request.Context(), eventIDStr, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position": entry.Position,
		"status":   entry.Status,
		"message":  "已加入等候名单，有人退票时将通知您",
	})
}

func (h *WaitlistHandler) GetWaitlistPosition(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	_, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	userID := strconv.FormatUint(uint64(userModel.ID), 10)

	entry, err := h.waitlistManager.GetWaitlistPosition(c.Request.Context(), eventIDStr, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position": entry.Position,
		"status":   entry.Status,
	})
}

func (h *WaitlistHandler) LeaveWaitlist(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	_, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	userID := strconv.FormatUint(uint64(userModel.ID), 10)

	if err := h.waitlistManager.LeaveWaitlist(c.Request.Context(), eventIDStr, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已离开等候名单"})
}
