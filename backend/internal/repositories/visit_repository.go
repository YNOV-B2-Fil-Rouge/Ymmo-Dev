package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type VisitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

// Create inserts a new visit.
func (r *VisitRepository) Create(v *models.Visit) error {
	return r.db.Create(v).Error
}

// FindByID returns a visit or nil.
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

// ListForUser returns the visits the user takes part in (as client or agent),
// soonest first.
func (r *VisitRepository) ListForUser(userID uint) ([]models.Visit, error) {
	var visits []models.Visit
	err := r.db.
		Where("client_id = ? OR agent_id = ?", userID, userID).
		Order("scheduled_at ASC").
		Find(&visits).Error
	return visits, err
}

// UpdateStatus changes a visit's status.
func (r *VisitRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Visit{}).Where("id = ?", id).
		Update("status", status).Error
}
