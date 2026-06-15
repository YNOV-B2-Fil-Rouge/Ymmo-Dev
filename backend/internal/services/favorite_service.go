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

func (s *FavoriteService) Add(userID, propertyID uint) error {
	property, err := s.properties.FindByID(propertyID)
	if err != nil {
		return err
	}
	if property == nil {
		return ErrPropertyNotFound
	}
	if !models.IsPubliclyVisible(property.Status) {
		return ErrPropertyNotPublic
	}
	return s.favorites.Add(userID, propertyID)
}

func (s *FavoriteService) Remove(userID, propertyID uint) error {
	return s.favorites.Remove(userID, propertyID)
}

func (s *FavoriteService) List(userID uint) ([]models.Property, error) {
	return s.favorites.ListProperties(userID)
}
