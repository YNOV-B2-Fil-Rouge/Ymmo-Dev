package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrAlertNotFound   = errors.New("alert not found")
	ErrNotAlertOwner   = errors.New("this alert does not belong to you")
	ErrEmptyAlert      = errors.New("an alert must define at least one criterion")
	ErrInvalidCategory = errors.New("unknown category")
	ErrDuplicateAlert  = errors.New("you already have an alert with these criteria")
)

type AlertService struct {
	alerts *repositories.AlertRepository
}

func NewAlertService(alerts *repositories.AlertRepository) *AlertService {
	return &AlertService{alerts: alerts}
}

func (s *AlertService) Create(userID uint, req dto.CreateAlertRequest) (*models.Alert, error) {
	if req.City == "" && req.CategoryID == nil && req.MinPrice == nil &&
		req.MaxPrice == nil && req.MinArea == nil && req.MaxEnergy == "" {
		return nil, ErrEmptyAlert
	}

	if req.CategoryID != nil {
		ok, err := s.alerts.CategoryExists(*req.CategoryID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrInvalidCategory
		}
	}

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

	dup, err := s.alerts.ExistsIdentical(alert)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, ErrDuplicateAlert
	}

	if err := s.alerts.Create(alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *AlertService) List(userID uint) ([]models.Alert, error) {
	return s.alerts.ListForUser(userID)
}

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
