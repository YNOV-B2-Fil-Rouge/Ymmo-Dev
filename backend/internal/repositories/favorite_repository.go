package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ymmo/internal/models"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Add links a property to a user. Idempotent: favoriting twice is a no-op
// (ON CONFLICT DO NOTHING) rather than a duplicate-key error.
func (r *FavoriteRepository) Add(userID, propertyID uint) error {
	fav := models.Favorite{UserID: userID, PropertyID: propertyID}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&fav).Error
}

// Remove unlinks a property from a user (no error if it was not a favorite).
func (r *FavoriteRepository) Remove(userID, propertyID uint) error {
	return r.db.Where("user_id = ? AND property_id = ?", userID, propertyID).
		Delete(&models.Favorite{}).Error
}

// ListProperties returns the user's favorited properties, newest first.
func (r *FavoriteRepository) ListProperties(userID uint) ([]models.Property, error) {
	var properties []models.Property
	err := r.db.
		Joins("JOIN favorites f ON f.property_id = properties.id").
		Where("f.user_id = ?", userID).
		Preload("Category").
		Preload("Photos").
		Order("f.created_at DESC").
		Find(&properties).Error
	return properties, err
}
