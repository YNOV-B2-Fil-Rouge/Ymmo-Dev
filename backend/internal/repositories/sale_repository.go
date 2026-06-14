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

func (r *SaleRepository) Create(s *models.SaleFile) error {
	return r.db.Create(s).Error
}

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

func (r *SaleRepository) ListForUser(userID uint) ([]models.SaleFile, error) {
	var sales []models.SaleFile
	err := r.db.
		Where("buyer_id = ? OR agent_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&sales).Error
	return sales, err
}

func (r *SaleRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.SaleFile{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SaleRepository) ExistsActive(propertyID, buyerID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.SaleFile{}).
		Where("property_id = ? AND buyer_id = ? AND status <> ?",
			propertyID, buyerID, models.SaleStatusCancelled).
		Count(&count).Error
	return count > 0, err
}
