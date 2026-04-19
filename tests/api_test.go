package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"event-booking/internal/database"
	"event-booking/internal/dto"
	"event-booking/internal/handlers"
	"event-booking/internal/jobs"
	"event-booking/internal/repositories"
	"event-booking/internal/router"
	"event-booking/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type queueRecorder struct {
	mu                   sync.Mutex
	bookingPayloads      []jobs.BookingConfirmationPayload
	eventUpdatedPayloads []jobs.EventUpdatedPayload
}

func (q *queueRecorder) EnqueueBookingConfirmation(_ context.Context, payload jobs.BookingConfirmationPayload) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.bookingPayloads = append(q.bookingPayloads, payload)
	return nil
}

func (q *queueRecorder) EnqueueEventUpdated(_ context.Context, payload jobs.EventUpdatedPayload) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.eventUpdatedPayloads = append(q.eventUpdatedPayloads, payload)
	return nil
}

func setupTestApp(t *testing.T) (*gin.Engine, *queueRecorder) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	queue := &queueRecorder{}
	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	authService := services.NewAuthService(userRepo, "test-secret", time.Hour)
	eventService := services.NewEventService(db, eventRepo, bookingRepo, queue)
	bookingService := services.NewBookingService(db, eventRepo, bookingRepo, queue)

	engine := router.New(router.Dependencies{
		AuthHandler:    handlers.NewAuthHandler(authService),
		EventHandler:   handlers.NewEventHandler(eventService),
		BookingHandler: handlers.NewBookingHandler(bookingService),
		HealthHandler:  handlers.NewHealthHandler(),
		AuthService:    authService,
	})

	return engine, queue
}

func TestEndToEndFlow(t *testing.T) {
	engine, queue := setupTestApp(t)

	organizerAuth := registerUser(t, engine, dto.RegisterRequest{
		Name:     "Org One",
		Email:    "organizer@example.com",
		Password: "password123",
		Role:     "organizer",
	})

	customerAuth := registerUser(t, engine, dto.RegisterRequest{
		Name:     "Customer One",
		Email:    "customer@example.com",
		Password: "password123",
		Role:     "customer",
	})

	eventBody := map[string]any{
		"title":         "Go Conference",
		"description":   "Backend event",
		"location":      "Bengaluru",
		"start_time":    "2026-05-01T10:00:00Z",
		"end_time":      "2026-05-01T18:00:00Z",
		"total_tickets": 10,
	}
	rec := performJSON(engine, http.MethodPost, "/events", eventBody, organizerAuth.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create event status=%d body=%s", rec.Code, rec.Body.String())
	}

	var eventResp dto.EventResponse
	decodeBody(t, rec, &eventResp)

	rec = performJSON(engine, http.MethodPost, "/events/"+itoa(eventResp.ID)+"/bookings", map[string]any{"ticket_count": 2}, customerAuth.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create booking status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = performJSON(engine, http.MethodPut, "/events/"+itoa(eventResp.ID), map[string]any{
		"title":         "Go Conference Updated",
		"description":   "Backend event updated",
		"location":      "Mumbai",
		"start_time":    "2026-05-02T10:00:00Z",
		"end_time":      "2026-05-02T18:00:00Z",
		"total_tickets": 12,
	}, organizerAuth.Token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update event status=%d body=%s", rec.Code, rec.Body.String())
	}

	if got := len(queue.bookingPayloads); got != 1 {
		t.Fatalf("expected 1 booking job, got %d", got)
	}
	if got := len(queue.eventUpdatedPayloads); got != 1 {
		t.Fatalf("expected 1 event update job, got %d", got)
	}
}

func TestDuplicateEmailRejected(t *testing.T) {
	engine, _ := setupTestApp(t)

	registerUser(t, engine, dto.RegisterRequest{
		Name:     "Org One",
		Email:    "repeat@example.com",
		Password: "password123",
		Role:     "organizer",
	})

	rec := performJSON(engine, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Org Two",
		Email:    "repeat@example.com",
		Password: "password123",
		Role:     "customer",
	}, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected conflict, got %d", rec.Code)
	}
}

func TestCustomerCannotCreateEvent(t *testing.T) {
	engine, _ := setupTestApp(t)

	customerAuth := registerUser(t, engine, dto.RegisterRequest{
		Name:     "Customer",
		Email:    "customer2@example.com",
		Password: "password123",
		Role:     "customer",
	})

	rec := performJSON(engine, http.MethodPost, "/events", map[string]any{
		"title":         "Nope",
		"description":   "Nope",
		"location":      "Nope",
		"start_time":    "2026-05-01T10:00:00Z",
		"end_time":      "2026-05-01T18:00:00Z",
		"total_tickets": 1,
	}, customerAuth.Token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", rec.Code)
	}
}

func TestOrganizerCannotBookTickets(t *testing.T) {
	engine, _ := setupTestApp(t)

	organizerAuth := registerUser(t, engine, dto.RegisterRequest{
		Name:     "Organizer",
		Email:    "organizer2@example.com",
		Password: "password123",
		Role:     "organizer",
	})

	eventResp := createEventForOrganizer(t, engine, organizerAuth.Token, "Organizer Event", 2)

	rec := performJSON(engine, http.MethodPost, "/events/"+itoa(eventResp.ID)+"/bookings", map[string]any{
		"ticket_count": 1,
	}, organizerAuth.Token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", rec.Code)
	}
}

func TestOrganizerCannotUpdateOthersEvent(t *testing.T) {
	engine, _ := setupTestApp(t)

	org1 := registerUser(t, engine, dto.RegisterRequest{Name: "Org1", Email: "org1@example.com", Password: "password123", Role: "organizer"})
	org2 := registerUser(t, engine, dto.RegisterRequest{Name: "Org2", Email: "org2@example.com", Password: "password123", Role: "organizer"})
	eventResp := createEventForOrganizer(t, engine, org1.Token, "Private Event", 5)

	rec := performJSON(engine, http.MethodPut, "/events/"+itoa(eventResp.ID), map[string]any{
		"title":         "Hacked",
		"description":   "No",
		"location":      "Delhi",
		"start_time":    "2026-05-03T10:00:00Z",
		"end_time":      "2026-05-03T12:00:00Z",
		"total_tickets": 5,
	}, org2.Token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCapacityAndOverbooking(t *testing.T) {
	engine, _ := setupTestApp(t)

	org := registerUser(t, engine, dto.RegisterRequest{Name: "Org", Email: "org3@example.com", Password: "password123", Role: "organizer"})
	customer := registerUser(t, engine, dto.RegisterRequest{Name: "Customer", Email: "cust3@example.com", Password: "password123", Role: "customer"})
	eventResp := createEventForOrganizer(t, engine, org.Token, "Limited Event", 2)

	rec := performJSON(engine, http.MethodPost, "/events/"+itoa(eventResp.ID)+"/bookings", map[string]any{"ticket_count": 2}, customer.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected created, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = performJSON(engine, http.MethodPost, "/events/"+itoa(eventResp.ID)+"/bookings", map[string]any{"ticket_count": 1}, customer.Token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateRejectsReducingBelowBookedCount(t *testing.T) {
	engine, _ := setupTestApp(t)

	org := registerUser(t, engine, dto.RegisterRequest{Name: "Org", Email: "org4@example.com", Password: "password123", Role: "organizer"})
	customer := registerUser(t, engine, dto.RegisterRequest{Name: "Customer", Email: "cust4@example.com", Password: "password123", Role: "customer"})
	eventResp := createEventForOrganizer(t, engine, org.Token, "Shrink Event", 5)

	rec := performJSON(engine, http.MethodPost, "/events/"+itoa(eventResp.ID)+"/bookings", map[string]any{"ticket_count": 4}, customer.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected created booking, got %d", rec.Code)
	}

	rec = performJSON(engine, http.MethodPut, "/events/"+itoa(eventResp.ID), map[string]any{
		"title":         "Shrink Event",
		"description":   "Desc",
		"location":      "Pune",
		"start_time":    "2026-06-01T10:00:00Z",
		"end_time":      "2026-06-01T12:00:00Z",
		"total_tickets": 3,
	}, org.Token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func registerUser(t *testing.T, engine *gin.Engine, request dto.RegisterRequest) dto.AuthResponse {
	t.Helper()
	rec := performJSON(engine, http.MethodPost, "/auth/register", request, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp dto.AuthResponse
	decodeBody(t, rec, &resp)
	return resp
}

func createEventForOrganizer(t *testing.T, engine *gin.Engine, token string, title string, totalTickets int) dto.EventResponse {
	t.Helper()
	rec := performJSON(engine, http.MethodPost, "/events", map[string]any{
		"title":         title,
		"description":   "Description",
		"location":      "Hyderabad",
		"start_time":    "2026-05-01T10:00:00Z",
		"end_time":      "2026-05-01T18:00:00Z",
		"total_tickets": totalTickets,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create event status=%d body=%s", rec.Code, rec.Body.String())
	}
	var eventResp dto.EventResponse
	decodeBody(t, rec, &eventResp)
	return eventResp
}

func performJSON(engine *gin.Engine, method, path string, payload any, token string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("decode body: %v body=%s", err, rec.Body.String())
	}
}

func itoa(id uint) string {
	return fmt.Sprintf("%d", id)
}
