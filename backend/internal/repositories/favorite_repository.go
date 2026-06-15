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

func (r *FavoriteRepository) Add(userID, propertyID uint) error {
	fav := models.Favorite{UserID: userID, PropertyID: propertyID}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&fav).Error
}

func (r *FavoriteRepository) Remove(userID, propertyID uint) error {
	return r.db.Where("user_id = ? AND property_id = ?", userID, propertyID).
		Delete(&models.Favorite{}).Error
}

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
