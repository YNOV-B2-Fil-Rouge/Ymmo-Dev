package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type PhotoRepository struct {
	db *gorm.DB
}

func NewPhotoRepository(db *gorm.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Add(photo *models.PropertyPhoto) error {
	return r.db.Create(photo).Error
}

func (r *PhotoRepository) FindByID(id uint) (*models.PropertyPhoto, error) {
	var photo models.PropertyPhoto
	err := r.db.First(&photo, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *PhotoRepository) Delete(id uint) error {
	return r.db.Delete(&models.PropertyPhoto{}, id).Error
}

func (r *PhotoRepository) CountByProperty(propertyID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.PropertyPhoto{}).
		Where("property_id = ?", propertyID).Count(&n).Error
	return n, err
}

func (r *PhotoRepository) ClearPrimary(propertyID uint) error {
	return r.db.Model(&models.PropertyPhoto{}).
		Where("property_id = ?", propertyID).
		Update("is_primary", false).Error
}
