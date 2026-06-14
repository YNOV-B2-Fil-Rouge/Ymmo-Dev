package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

// SellerApplicationRepository handles persistence for buyer→seller requests.
type SellerApplicationRepository struct {
	db *gorm.DB
}

func NewSellerApplicationRepository(db *gorm.DB) *SellerApplicationRepository {
	return &SellerApplicationRepository{db: db}
}

// Create inserts a new application.
func (r *SellerApplicationRepository) Create(app *models.SellerApplication) error {
	return r.db.Create(app).Error
}

// FindByID returns one application (with its applicant) or nil.
func (r *SellerApplicationRepository) FindByID(id uint) (*models.SellerApplication, error) {
	var app models.SellerApplication
	err := r.db.Preload("User").First(&app, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// FindLatestByUser returns the user's most recent application, or nil.
func (r *SellerApplicationRepository) FindLatestByUser(userID uint) (*models.SellerApplication, error) {
	var app models.SellerApplication
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// HasPending reports whether the user already has an unresolved request.
func (r *SellerApplicationRepository) HasPending(userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.SellerApplication{}).
		Where("user_id = ? AND status = ?", userID, models.SellerApplicationPending).
		Count(&count).Error
	return count > 0, err
}

// ListPending returns every application awaiting review (newest first).
func (r *SellerApplicationRepository) ListPending() ([]models.SellerApplication, error) {
	var apps []models.SellerApplication
	err := r.db.
		Where("status = ?", models.SellerApplicationPending).
		Preload("User").
		Order("created_at DESC").
		Find(&apps).Error
	return apps, err
}

// Review sets the outcome (APPROVED/REJECTED), the reviewer and the timestamp.
func (r *SellerApplicationRepository) Review(id uint, status string, reviewerID uint) error {
	return r.db.Model(&models.SellerApplication{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"reviewed_by": reviewerID,
			"reviewed_at": time.Now(),
		}).Error
}
