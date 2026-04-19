package services

import (
	"context"
	"errors"

	"event-booking/internal/dto"
	"event-booking/internal/jobs"
	"event-booking/internal/models"
	"event-booking/internal/repositories"

	"gorm.io/gorm"
)

type EventService struct {
	db       *gorm.DB
	events   *repositories.EventRepository
	bookings *repositories.BookingRepository
	queue    jobs.Enqueuer
}

func NewEventService(db *gorm.DB, events *repositories.EventRepository, bookings *repositories.BookingRepository, queue jobs.Enqueuer) *EventService {
	return &EventService{
		db:       db,
		events:   events,
		bookings: bookings,
		queue:    queue,
	}
}

func (s *EventService) Create(organizerID uint, input dto.EventCreateRequest) (*models.Event, error) {
	if !input.StartTime.Before(input.EndTime) {
		return nil, ErrInvalidEventTime
	}

	event := &models.Event{
		OrganizerID:      organizerID,
		Title:            input.Title,
		Description:      input.Description,
		Location:         input.Location,
		StartTime:        input.StartTime,
		EndTime:          input.EndTime,
		TotalTickets:     input.TotalTickets,
		AvailableTickets: input.TotalTickets,
	}

	if err := s.events.Create(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) Update(ctx context.Context, organizerID uint, eventID uint, input dto.EventUpdateRequest) (*models.Event, error) {
	if !input.StartTime.Before(input.EndTime) {
		return nil, ErrInvalidEventTime
	}

	var updated *models.Event
	err := s.db.Transaction(func(tx *gorm.DB) error {
		event, err := s.events.FindByIDForUpdate(tx, eventID)
		if err != nil {
			return err
		}
		if event.OrganizerID != organizerID {
			return ErrForbidden
		}

		booked, err := s.bookings.CountTicketsForEvent(tx, eventID)
		if err != nil {
			return err
		}
		if int(booked) > input.TotalTickets {
			return ErrCapacityTooLow
		}

		event.Title = input.Title
		event.Description = input.Description
		event.Location = input.Location
		event.StartTime = input.StartTime
		event.EndTime = input.EndTime
		event.TotalTickets = input.TotalTickets
		event.AvailableTickets = input.TotalTickets - int(booked)

		if err := s.events.Update(tx, event); err != nil {
			return err
		}

		updated = event
		return nil
	})
	if err != nil {
		return nil, err
	}

	customers, err := s.bookings.DistinctCustomersForEvent(eventID)
	if err != nil {
		return nil, err
	}

	for _, customer := range customers {
		err := s.queue.EnqueueEventUpdated(ctx, jobs.EventUpdatedPayload{
			EventID:       updated.ID,
			EventTitle:    updated.Title,
			CustomerName:  customer.Name,
			CustomerEmail: customer.Email,
		})
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *EventService) ListAll() ([]models.Event, error) {
	return s.events.ListAll()
}

func (s *EventService) ListByOrganizer(organizerID uint) ([]models.Event, error) {
	return s.events.ListByOrganizer(organizerID)
}

func (s *EventService) GetByID(eventID uint) (*models.Event, error) {
	return s.events.FindByID(eventID)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
