package repositories

import (
	"event-booking/internal/models"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(tx *gorm.DB, booking *models.Booking) error {
	return tx.Create(booking).Error
}

func (r *BookingRepository) ListByCustomer(customerID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	if err := r.db.Preload("Event").Where("customer_id = ?", customerID).Order("created_at desc").Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *BookingRepository) CountTicketsForEvent(tx *gorm.DB, eventID uint) (int64, error) {
	var total int64
	if err := tx.Model(&models.Booking{}).Where("event_id = ?", eventID).Select("COALESCE(SUM(ticket_count), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *BookingRepository) DistinctCustomersForEvent(eventID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Model(&models.User{}).
		Joins("join bookings on bookings.customer_id = users.id").
		Where("bookings.event_id = ?", eventID).
		Distinct("users.id", "users.name", "users.email", "users.role", "users.created_at", "users.updated_at").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
