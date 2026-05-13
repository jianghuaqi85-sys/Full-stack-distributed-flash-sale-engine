package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type ShowHandler struct {
	showService *service.ShowService
}

func NewShowHandler(showService *service.ShowService) *ShowHandler {
	return &ShowHandler{showService: showService}
}

func (h *ShowHandler) CreateShow(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		ShowTime  string `json:"show_time" binding:"required"`
		EndTime   string `json:"end_time" binding:"required"`
		Stock     int    `json:"stock" binding:"required,min=0"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showTime, err := time.Parse(time.RFC3339, req.ShowTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show_time format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	show, err := h.showService.CreateShow(service.CreateShowInput{
		EventID:   uint(eventID),
		Name:      req.Name,
		ShowTime:  showTime,
		EndTime:   endTime,
		Stock:     req.Stock,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         show.ID,
		"name":       show.Name,
		"show_time":  show.ShowTime,
		"end_time":   show.EndTime,
		"status":     show.Status,
		"stock":      show.Stock,
		"sort_order": show.SortOrder,
	})
}

func (h *ShowHandler) UpdateShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		ShowTime  string `json:"show_time" binding:"required"`
		EndTime   string `json:"end_time" binding:"required"`
		Stock     int    `json:"stock" binding:"required,min=0"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showTime, err := time.Parse(time.RFC3339, req.ShowTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show_time format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	show, err := h.showService.UpdateShow(uint(id), service.UpdateShowInput{
		Name:      req.Name,
		ShowTime:  showTime,
		EndTime:   endTime,
		Stock:     req.Stock,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         show.ID,
		"name":       show.Name,
		"show_time":  show.ShowTime,
		"end_time":   show.EndTime,
		"status":     show.Status,
		"stock":      show.Stock,
		"sort_order": show.SortOrder,
	})
}

func (h *ShowHandler) DeleteShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	if err := h.showService.DeleteShow(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ShowHandler) PublishShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	if err := h.showService.PublishShow(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "场次已上架"})
}

func (h *ShowHandler) UnpublishShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	if err := h.showService.UnpublishShow(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "场次已下架"})
}

func (h *ShowHandler) ListShows(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	shows, err := h.showService.ListShowsByEvent(uint(eventID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shows})
}

func (h *ShowHandler) GetShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	show, err := h.showService.GetShow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, show)
}
