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

// Add inserts a new photo row.
func (r *PhotoRepository) Add(photo *models.PropertyPhoto) error {
	return r.db.Create(photo).Error
}

// FindByID returns a photo or nil if it does not exist.
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

// Delete removes a photo by id.
func (r *PhotoRepository) Delete(id uint) error {
	return r.db.Delete(&models.PropertyPhoto{}, id).Error
}

// CountByProperty counts how many photos a property already has
// (used to set the new photo's sort order).
func (r *PhotoRepository) CountByProperty(propertyID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.PropertyPhoto{}).
		Where("property_id = ?", propertyID).Count(&n).Error
	return n, err
}

// ClearPrimary unsets the primary flag on every photo of a property,
// so a newly set primary is the only one.
func (r *PhotoRepository) ClearPrimary(propertyID uint) error {
	return r.db.Model(&models.PropertyPhoto{}).
		Where("property_id = ?", propertyID).
		Update("is_primary", false).Error
}
