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

// CategoryExists reports whether a property category id exists (so we reject
// a bad category with a clean 400 instead of hitting a FK 500).
func (r *AlertRepository) CategoryExists(id uint8) (bool, error) {
	var count int64
	err := r.db.Table("property_categories").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsIdentical reports whether the SAME user already has an alert with the
// exact same criteria. `<=>` is MariaDB's NULL-safe equality, so unset
// (NULL) criteria are compared correctly. Scoped to the user, so a different
// user may keep an identical alert.
func (r *AlertRepository) ExistsIdentical(a *models.Alert) (bool, error) {
	var count int64
	err := r.db.Model(&models.Alert{}).
		Where("user_id = ?", a.UserID).
		Where("city <=> ?", a.City).
		Where("category_id <=> ?", a.CategoryID).
		Where("min_price <=> ?", a.MinPrice).
		Where("max_price <=> ?", a.MaxPrice).
		Where("min_area <=> ?", a.MinArea).
		Where("max_energy <=> ?", a.MaxEnergy).
		Count(&count).Error
	return count > 0, err
}
