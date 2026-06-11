package services

import (
	"errors"
	"time"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrVisitNotFound      = errors.New("visit not found")
	ErrNotVisitParticipant = errors.New("not a participant of this visit")
	ErrNoAgentForProperty = errors.New("this property has no assigned agent")
	ErrPastSchedule       = errors.New("the scheduled date must be in the future")
)

type VisitService struct {
	visits     *repositories.VisitRepository
	properties *repositories.PropertyRepository
}

func NewVisitService(visits *repositories.VisitRepository, properties *repositories.PropertyRepository) *VisitService {
	return &VisitService{visits: visits, properties: properties}
}

// Request creates a visit request for a property. The agent is taken from the
// property; the current user is the client.
func (s *VisitService) Request(clientID, propertyID uint, req dto.RequestVisitRequest) (*models.Visit, error) {
	if req.ScheduledAt.Before(time.Now()) {
		return nil, ErrPastSchedule
	}

	property, err := s.properties.FindByID(propertyID)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, ErrPropertyNotFound
	}
	if property.AgentID == nil {
		return nil, ErrNoAgentForProperty
	}

	visit := &models.Visit{
		PropertyID:  propertyID,
		ClientID:    clientID,
		AgentID:     *property.AgentID,
		ScheduledAt: req.ScheduledAt,
		Status:      models.VisitStatusRequested,
	}
	if req.Notes != "" {
		visit.Notes = &req.Notes
	}
	if err := s.visits.Create(visit); err != nil {
		return nil, err
	}
	return visit, nil
}

// List returns the user's visits.
func (s *VisitService) List(userID uint) ([]models.Visit, error) {
	return s.visits.ListForUser(userID)
}

// UpdateStatus changes a visit's status, restricted to its participants.
func (s *VisitService) UpdateStatus(userID, visitID uint, status string) (*models.Visit, error) {
	visit, err := s.visits.FindByID(visitID)
	if err != nil {
		return nil, err
	}
	if visit == nil {
		return nil, ErrVisitNotFound
	}
	if visit.ClientID != userID && visit.AgentID != userID {
		return nil, ErrNotVisitParticipant
	}
	if err := s.visits.UpdateStatus(visitID, status); err != nil {
		return nil, err
	}
	return s.visits.FindByID(visitID)
}
