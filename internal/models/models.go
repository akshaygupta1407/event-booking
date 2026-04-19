package models

import "time"

const (
	RoleOrganizer = "organizer"
	RoleCustomer  = "customer"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:20;not null;index" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Event struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	OrganizerID      uint      `gorm:"not null;index" json:"organizer_id"`
	Organizer        User      `json:"organizer,omitempty"`
	Title            string    `gorm:"size:200;not null" json:"title"`
	Description      string    `gorm:"type:text;not null" json:"description"`
	Location         string    `gorm:"size:255;not null" json:"location"`
	StartTime        time.Time `gorm:"not null" json:"start_time"`
	EndTime          time.Time `gorm:"not null" json:"end_time"`
	TotalTickets     int       `gorm:"not null" json:"total_tickets"`
	AvailableTickets int       `gorm:"not null" json:"available_tickets"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Booking struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EventID     uint      `gorm:"not null;index" json:"event_id"`
	Event       Event     `json:"event,omitempty"`
	CustomerID  uint      `gorm:"not null;index" json:"customer_id"`
	Customer    User      `json:"customer,omitempty"`
	TicketCount int       `gorm:"not null" json:"ticket_count"`
	CreatedAt   time.Time `json:"created_at"`
}
