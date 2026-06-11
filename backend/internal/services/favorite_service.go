package services

import (
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

type FavoriteService struct {
	favorites  *repositories.FavoriteRepository
	properties *repositories.PropertyRepository
}

func NewFavoriteService(favorites *repositories.FavoriteRepository, properties *repositories.PropertyRepository) *FavoriteService {
	return &FavoriteService{favorites: favorites, properties: properties}
}

// Add favorites a property for a user, after checking the property exists.
// Reuses ErrPropertyNotFound (defined in property_service.go).
func (s *FavoriteService) Add(userID, propertyID uint) error {
	property, err := s.properties.FindByID(propertyID)
	if err != nil {
		return err
	}
	if property == nil {
		return ErrPropertyNotFound
	}
	return s.favorites.Add(userID, propertyID)
}

// Remove un-favorites a property for a user.
func (s *FavoriteService) Remove(userID, propertyID uint) error {
	return s.favorites.Remove(userID, propertyID)
}

// List returns the user's favorited properties.
func (s *FavoriteService) List(userID uint) ([]models.Property, error) {
	return s.favorites.ListProperties(userID)
}
