package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// Create inserts a new sale file.
func (r *SaleRepository) Create(s *models.SaleFile) error {
	return r.db.Create(s).Error
}

// FindByID returns a sale file or nil.
func (r *SaleRepository) FindByID(id uint) (*models.SaleFile, error) {
	var s models.SaleFile
	err := r.db.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListForUser returns the sale files the user is involved in (as buyer or
// agent), newest first.
func (r *SaleRepository) ListForUser(userID uint) ([]models.SaleFile, error) {
	var sales []models.SaleFile
	err := r.db.
		Where("buyer_id = ? OR agent_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&sales).Error
	return sales, err
}

// Update applies a partial set of columns to a sale file.
func (r *SaleRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.SaleFile{}).Where("id = ?", id).Updates(updates).Error
}
