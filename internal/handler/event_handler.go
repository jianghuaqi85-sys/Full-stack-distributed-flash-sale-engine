package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Location    string `json:"location" binding:"required"`
		CoverImage  string `json:"cover_image"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use RFC3339"})
		return
	}

	event, err := h.eventService.CreateEvent(service.CreateEventInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		CoverImage:  req.CoverImage,
		StartTime:   startTime,
		EndTime:     endTime,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         event.ID,
		"title":      event.Title,
		"location":   event.Location,
		"start_time": event.StartTime,
		"end_time":   event.EndTime,
		"status":     event.Status,
	})
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Location    string `json:"location" binding:"required"`
		CoverImage  string `json:"cover_image"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use RFC3339"})
		return
	}

	event, err := h.eventService.UpdateEvent(uint(id), service.UpdateEventInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		CoverImage:  req.CoverImage,
		StartTime:   startTime,
		EndTime:     endTime,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         event.ID,
		"title":      event.Title,
		"location":   event.Location,
		"start_time": event.StartTime,
		"end_time":   event.EndTime,
		"status":     event.Status,
	})
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.eventService.GetEvent(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	events, total, err := h.eventService.ListEvents(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  events,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *EventHandler) PublishEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	if err := h.eventService.PublishEvent(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event published"})
}

func (h *EventHandler) UnpublishEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	if err := h.eventService.UnpublishEvent(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event unpublished"})
}

func (h *EventHandler) CreateTicketType(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req struct {
		Name       string  `json:"name" binding:"required"`
		Price      float64 `json:"price" binding:"required,min=0"`
		Stock      int     `json:"stock" binding:"required,min=0"`
		MaxPerUser int     `json:"max_per_user"`
		SortOrder  int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tt, err := h.eventService.CreateTicketType(uint(eventID), service.CreateTicketTypeInput{
		Name:       req.Name,
		Price:      req.Price,
		Stock:      req.Stock,
		MaxPerUser: req.MaxPerUser,
		SortOrder:  req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          tt.ID,
		"name":        tt.Name,
		"price":       tt.Price,
		"stock":       tt.Stock,
		"max_per_user": tt.MaxPerUser,
	})
}

func (h *EventHandler) UpdateTicketType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket type id"})
		return
	}

	var req struct {
		Name       string  `json:"name" binding:"required"`
		Price      float64 `json:"price" binding:"required,min=0"`
		Stock      int     `json:"stock" binding:"required,min=0"`
		MaxPerUser int     `json:"max_per_user"`
		SortOrder  int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tt, err := h.eventService.UpdateTicketType(uint(id), service.CreateTicketTypeInput{
		Name:       req.Name,
		Price:      req.Price,
		Stock:      req.Stock,
		MaxPerUser: req.MaxPerUser,
		SortOrder:  req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          tt.ID,
		"name":        tt.Name,
		"price":       tt.Price,
		"stock":       tt.Stock,
		"max_per_user": tt.MaxPerUser,
	})
}

func (h *EventHandler) DeleteTicketType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket type id"})
		return
	}

	if err := h.eventService.DeleteTicketType(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *EventHandler) GetEventStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	stock, err := h.eventService.GetEventStock(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stock)
}

func (h *EventHandler) EndEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	if err := h.eventService.EndEvent(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event ended"})
}
