package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type VisitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

func (r *VisitRepository) Create(v *models.Visit) error {
	return r.db.Create(v).Error
}

func (r *VisitRepository) FindByID(id uint) (*models.Visit, error) {
	var v models.Visit
	err := r.db.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VisitRepository) ListForUser(userID uint) ([]models.Visit, error) {
	var visits []models.Visit
	err := r.db.
		Where("client_id = ? OR agent_id = ?", userID, userID).
		Order("scheduled_at ASC").
		Find(&visits).Error
	return visits, err
}

func (r *VisitRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Visit{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *VisitRepository) ExistsForSchedule(propertyID, clientID uint, scheduledAt time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.Visit{}).
		Where("property_id = ? AND client_id = ? AND scheduled_at = ? AND status <> ?",
			propertyID, clientID, scheduledAt, models.VisitStatusCancelled).
		Count(&count).Error
	return count > 0, err
}
