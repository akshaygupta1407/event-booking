package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Role     string `json:"role" binding:"required,oneof=organizer customer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserPayload `json:"user"`
}

type UserPayload struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type EventCreateRequest struct {
	Title        string    `json:"title" binding:"required,min=3,max=200"`
	Description  string    `json:"description" binding:"required"`
	Location     string    `json:"location" binding:"required,min=2,max=255"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	TotalTickets int       `json:"total_tickets" binding:"required,gt=0"`
}

type EventUpdateRequest struct {
	Title        string    `json:"title" binding:"required,min=3,max=200"`
	Description  string    `json:"description" binding:"required"`
	Location     string    `json:"location" binding:"required,min=2,max=255"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	TotalTickets int       `json:"total_tickets" binding:"required,gt=0"`
}

type EventResponse struct {
	ID               uint      `json:"id"`
	OrganizerID      uint      `json:"organizer_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Location         string    `json:"location"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	TotalTickets     int       `json:"total_tickets"`
	AvailableTickets int       `json:"available_tickets"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type BookingCreateRequest struct {
	TicketCount int `json:"ticket_count" binding:"required,gt=0"`
}

type BookingResponse struct {
	ID          uint          `json:"id"`
	EventID     uint          `json:"event_id"`
	CustomerID  uint          `json:"customer_id"`
	TicketCount int           `json:"ticket_count"`
	CreatedAt   time.Time     `json:"created_at"`
	Event       EventResponse `json:"event"`
}
