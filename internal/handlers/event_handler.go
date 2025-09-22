package handlers

import (
	"net/http"
	"strconv"

	"highload-microservice/internal/models"
	"highload-microservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type EventHandler struct {
	eventService *services.EventService
	logger       *logrus.Logger
}

func NewEventHandler(eventService *services.EventService, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		logger:       logger,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	event, err := h.eventService.CreateEvent(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to create event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Errorf("Invalid event ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventService.GetEvent(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("Failed to get event: %v", err)
		if err.Error() == "event not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	events, err := h.eventService.ListEvents(c.Request.Context(), page, limit)
	if err != nil {
		h.logger.Errorf("Failed to list events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	c.JSON(http.StatusOK, events)
}
