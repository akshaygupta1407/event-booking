package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"event-booking/internal/dto"
	"event-booking/internal/middleware"
	"event-booking/internal/models"
	"event-booking/internal/services"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) Create(c *gin.Context) {
	var input dto.EventCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := middleware.MustUser(c)
	event, err := h.service.Create(user.ID, input)
	if err != nil {
		handleEventError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toEventResponse(*event))
}

func (h *EventHandler) Update(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var input dto.EventUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := middleware.MustUser(c)
	event, err := h.service.Update(c.Request.Context(), user.ID, uint(eventID), input)
	if err != nil {
		handleEventError(c, err)
		return
	}

	c.JSON(http.StatusOK, toEventResponse(*event))
}

func (h *EventHandler) ListAll(c *gin.Context) {
	events, err := h.service.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	response := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, toEventResponse(event))
	}
	c.JSON(http.StatusOK, response)
}

func (h *EventHandler) GetByID(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.service.GetByID(uint(eventID))
	if err != nil {
		handleEventError(c, err)
		return
	}

	c.JSON(http.StatusOK, toEventResponse(*event))
}

func (h *EventHandler) ListOrganizerEvents(c *gin.Context) {
	user := middleware.MustUser(c)
	events, err := h.service.ListByOrganizer(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizer events"})
		return
	}

	response := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, toEventResponse(event))
	}
	c.JSON(http.StatusOK, response)
}

func handleEventError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidEventTime), errors.Is(err, services.ErrCapacityTooLow):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case services.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "event request failed"})
	}
}

func toEventResponse(event models.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:               event.ID,
		OrganizerID:      event.OrganizerID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		StartTime:        event.StartTime,
		EndTime:          event.EndTime,
		TotalTickets:     event.TotalTickets,
		AvailableTickets: event.AvailableTickets,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}
}
