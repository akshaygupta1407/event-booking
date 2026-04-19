package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

const (
	TypeBookingConfirmation = "booking:confirmation"
	TypeEventUpdated        = "event:updated"
)

type BookingConfirmationPayload struct {
	BookingID     uint   `json:"booking_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	EventTitle    string `json:"event_title"`
	TicketCount   int    `json:"ticket_count"`
}

type EventUpdatedPayload struct {
	EventID       uint   `json:"event_id"`
	EventTitle    string `json:"event_title"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}

type Enqueuer interface {
	EnqueueBookingConfirmation(ctx context.Context, payload BookingConfirmationPayload) error
	EnqueueEventUpdated(ctx context.Context, payload EventUpdatedPayload) error
}

type AsynqQueue struct {
	client *asynq.Client
}

func NewAsynqQueue(redisAddress string) *AsynqQueue {
	return &AsynqQueue{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddress}),
	}
}

func (q *AsynqQueue) EnqueueBookingConfirmation(ctx context.Context, payload BookingConfirmationPayload) error {
	return q.enqueue(ctx, TypeBookingConfirmation, payload)
}

func (q *AsynqQueue) EnqueueEventUpdated(ctx context.Context, payload EventUpdatedPayload) error {
	return q.enqueue(ctx, TypeEventUpdated, payload)
}

func (q *AsynqQueue) enqueue(ctx context.Context, taskType string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = q.client.EnqueueContext(ctx, asynq.NewTask(taskType, body))
	return err
}

func (q *AsynqQueue) Close() error {
	return q.client.Close()
}

func NewServer(redisAddress string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddress},
		asynq.Config{
			Concurrency: 10,
		},
	)
}

func NewMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeBookingConfirmation, handleBookingConfirmation)
	mux.HandleFunc(TypeEventUpdated, handleEventUpdated)
	return mux
}

func handleBookingConfirmation(_ context.Context, task *asynq.Task) error {
	var payload BookingConfirmationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal booking confirmation payload: %w", err)
	}

	log.Printf("booking confirmation email: booking_id=%d customer=%s email=%s event=%s tickets=%d",
		payload.BookingID, payload.CustomerName, payload.CustomerEmail, payload.EventTitle, payload.TicketCount)
	return nil
}

func handleEventUpdated(_ context.Context, task *asynq.Task) error {
	var payload EventUpdatedPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal event updated payload: %w", err)
	}

	log.Printf("event update notification: event_id=%d event=%s customer=%s email=%s",
		payload.EventID, payload.EventTitle, payload.CustomerName, payload.CustomerEmail)
	return nil
}
