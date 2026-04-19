package services

import (
	"context"

	"event-booking/internal/dto"
	"event-booking/internal/jobs"
	"event-booking/internal/models"
	"event-booking/internal/repositories"

	"gorm.io/gorm"
)

type BookingService struct {
	db       *gorm.DB
	events   *repositories.EventRepository
	bookings *repositories.BookingRepository
	queue    jobs.Enqueuer
}

func NewBookingService(db *gorm.DB, events *repositories.EventRepository, bookings *repositories.BookingRepository, queue jobs.Enqueuer) *BookingService {
	return &BookingService{
		db:       db,
		events:   events,
		bookings: bookings,
		queue:    queue,
	}
}

func (s *BookingService) Create(ctx context.Context, customer dto.UserPayload, eventID uint, input dto.BookingCreateRequest) (*models.Booking, error) {
	var booking *models.Booking
	var event *models.Event

	err := s.db.Transaction(func(tx *gorm.DB) error {
		lockedEvent, err := s.events.FindByIDForUpdate(tx, eventID)
		if err != nil {
			return err
		}
		if lockedEvent.AvailableTickets < input.TicketCount {
			return ErrEventCapacity
		}

		lockedEvent.AvailableTickets -= input.TicketCount
		if err := s.events.Update(tx, lockedEvent); err != nil {
			return err
		}

		created := &models.Booking{
			EventID:     eventID,
			CustomerID:  customer.ID,
			TicketCount: input.TicketCount,
		}
		if err := s.bookings.Create(tx, created); err != nil {
			return err
		}

		booking = created
		event = lockedEvent
		return nil
	})
	if err != nil {
		return nil, err
	}

	booking.Event = *event
	if err := s.queue.EnqueueBookingConfirmation(ctx, jobs.BookingConfirmationPayload{
		BookingID:     booking.ID,
		CustomerName:  customer.Name,
		CustomerEmail: customer.Email,
		EventTitle:    event.Title,
		TicketCount:   booking.TicketCount,
	}); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) ListByCustomer(customerID uint) ([]models.Booking, error) {
	return s.bookings.ListByCustomer(customerID)
}
