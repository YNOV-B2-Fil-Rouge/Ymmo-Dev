package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var ErrPhotoNotFound = errors.New("photo not found")

type PhotoService struct {
	photos     *repositories.PhotoRepository
	properties *repositories.PropertyRepository
}

func NewPhotoService(photos *repositories.PhotoRepository, properties *repositories.PropertyRepository) *PhotoService {
	return &PhotoService{photos: photos, properties: properties}
}

func (s *PhotoService) Add(propertyID uint, req dto.AddPhotoRequest) (*models.PropertyPhoto, error) {
	property, err := s.properties.FindByID(propertyID)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, ErrPropertyNotFound
	}

	if req.IsPrimary {
		if err := s.photos.ClearPrimary(propertyID); err != nil {
			return nil, err
		}
	}

	count, err := s.photos.CountByProperty(propertyID)
	if err != nil {
		return nil, err
	}

	photo := &models.PropertyPhoto{
		PropertyID: propertyID,
		URL:        req.URL,
		SortOrder:  uint8(count),
		IsPrimary:  req.IsPrimary,
	}
	if err := s.photos.Add(photo); err != nil {
		return nil, err
	}
	return photo, nil
}

func (s *PhotoService) Delete(photoID uint) error {
	photo, err := s.photos.FindByID(photoID)
	if err != nil {
		return err
	}
	if photo == nil {
		return ErrPhotoNotFound
	}
	return s.photos.Delete(photoID)
}
