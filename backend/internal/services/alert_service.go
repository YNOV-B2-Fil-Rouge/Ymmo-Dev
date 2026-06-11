package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrAlertNotFound = errors.New("alert not found")
	ErrNotAlertOwner = errors.New("this alert does not belong to you")
)

type AlertService struct {
	alerts *repositories.AlertRepository
}

func NewAlertService(alerts *repositories.AlertRepository) *AlertService {
	return &AlertService{alerts: alerts}
}

// Create saves a new active alert for the user.
func (s *AlertService) Create(userID uint, req dto.CreateAlertRequest) (*models.Alert, error) {
	alert := &models.Alert{
		UserID:     userID,
		CategoryID: req.CategoryID,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		MinArea:    req.MinArea,
		IsActive:   true,
	}
	if req.City != "" {
		alert.City = &req.City
	}
	if req.MaxEnergy != "" {
		alert.MaxEnergy = &req.MaxEnergy
	}

	if err := s.alerts.Create(alert); err != nil {
		return nil, err
	}
	return alert, nil
}

// List returns the user's alerts.
func (s *AlertService) List(userID uint) ([]models.Alert, error) {
	return s.alerts.ListForUser(userID)
}

// Delete removes an alert, but only if it belongs to the user.
func (s *AlertService) Delete(userID, alertID uint) error {
	alert, err := s.alerts.FindByID(alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return ErrAlertNotFound
	}
	if alert.UserID != userID {
		return ErrNotAlertOwner
	}
	return s.alerts.Delete(alertID)
}
