package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/pkg/ws"
	"order-system/internal/queue"
)

type QueueHandler struct {
	queueManager *queue.QueueManager
	wsHub        *ws.Hub
}

func NewQueueHandler(queueManager *queue.QueueManager, wsHub *ws.Hub) *QueueHandler {
	return &QueueHandler{
		queueManager: queueManager,
		wsHub:        wsHub,
	}
}

func (h *QueueHandler) JoinQueue(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	userModel, ok := getUser(c)
	if !ok {
		return
	}

	userID := strconv.FormatUint(uint64(userModel.ID), 10)
	eventIDStr = strconv.FormatUint(eventID, 10)

	position, err := h.queueManager.JoinQueue(c.Request.Context(), eventIDStr, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position":      position.Position,
		"total_ahead":   position.TotalAhead,
		"estimated_wait": position.EstimatedWait.Seconds(),
		"status":        position.Status,
	})
}

func (h *QueueHandler) GetPosition(c *gin.Context) {
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

	position, err := h.queueManager.GetPosition(c.Request.Context(), eventIDStr, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position":      position.Position,
		"total_ahead":   position.TotalAhead,
		"estimated_wait": position.EstimatedWait.Seconds(),
		"status":        position.Status,
	})
}

func (h *QueueHandler) LeaveQueue(c *gin.Context) {
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

	if err := h.queueManager.LeaveQueue(c.Request.Context(), eventIDStr, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已离开队列"})
}
