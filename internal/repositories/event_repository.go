package repositories

import (
	"event-booking/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*models.Event, error) {
	var event models.Event
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) ListAll() ([]models.Event, error) {
	var events []models.Event
	if err := r.db.Order("start_time asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) ListByOrganizer(organizerID uint) ([]models.Event, error) {
	var events []models.Event
	if err := r.db.Where("organizer_id = ?", organizerID).Order("start_time asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) Update(tx *gorm.DB, event *models.Event) error {
	return tx.Save(event).Error
}
