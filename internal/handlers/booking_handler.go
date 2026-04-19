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

type BookingHandler struct {
	service *services.BookingService
}

func NewBookingHandler(service *services.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) Create(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var input dto.BookingCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := middleware.MustUser(c)
	booking, err := h.service.Create(c.Request.Context(), user, uint(eventID), input)
	if err != nil {
		handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toBookingResponse(*booking))
}

func (h *BookingHandler) ListMine(c *gin.Context) {
	user := middleware.MustUser(c)
	bookings, err := h.service.ListByCustomer(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	response := make([]dto.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		response = append(response, toBookingResponse(booking))
	}
	c.JSON(http.StatusOK, response)
}

func handleBookingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrEventCapacity):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case services.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking request failed"})
	}
}

func toBookingResponse(booking models.Booking) dto.BookingResponse {
	return dto.BookingResponse{
		ID:          booking.ID,
		EventID:     booking.EventID,
		CustomerID:  booking.CustomerID,
		TicketCount: booking.TicketCount,
		CreatedAt:   booking.CreatedAt,
		Event:       toEventResponse(booking.Event),
	}
}
