package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create inserts a new alert.
func (r *AlertRepository) Create(a *models.Alert) error {
	return r.db.Create(a).Error
}

// FindByID returns an alert or nil.
func (r *AlertRepository) FindByID(id uint) (*models.Alert, error) {
	var a models.Alert
	err := r.db.First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListForUser returns a user's alerts, newest first.
func (r *AlertRepository) ListForUser(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// Delete removes an alert by id.
func (r *AlertRepository) Delete(id uint) error {
	return r.db.Delete(&models.Alert{}, id).Error
}
