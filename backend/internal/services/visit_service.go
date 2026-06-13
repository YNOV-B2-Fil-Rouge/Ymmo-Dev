package services

import (
	"errors"
	"time"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrVisitNotFound       = errors.New("visit not found")
	ErrNotVisitParticipant = errors.New("not a participant of this visit")
	ErrNoAgentForProperty  = errors.New("this property has no assigned agent")
	ErrPastSchedule        = errors.New("the scheduled date must be in the future")
	// ErrVisitForbidden: action not allowed for this user (e.g. an agent
	// requesting a visit, or a buyer trying to confirm one).
	ErrVisitForbidden       = errors.New("you are not allowed to perform this action on the visit")
	ErrPropertyNotAvailable = errors.New("this property is not available for visits")
	ErrDuplicateVisit       = errors.New("a visit is already requested for this property at this date")
)

type VisitService struct {
	visits     *repositories.VisitRepository
	properties *repositories.PropertyRepository
}

func NewVisitService(visits *repositories.VisitRepository, properties *repositories.PropertyRepository) *VisitService {
	return &VisitService{visits: visits, properties: properties}
}

// Request creates a visit request for a property. Only a client (BUYER/SELLER)
// may request, on an AVAILABLE property, and not twice for the same date.
func (s *VisitService) Request(clientID uint, clientRole string, propertyID uint, req dto.RequestVisitRequest) (*models.Visit, error) {
	// Only clients request visits — never an agent/director/HQ/IT.
	if clientRole != "BUYER" && clientRole != "SELLER" {
		return nil, ErrVisitForbidden
	}
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
	// Only published properties can be visited (not drafts/sold/withdrawn).
	if property.Status != models.PropertyStatusAvailable {
		return nil, ErrPropertyNotAvailable
	}
	// You cannot request a visit for your own listing.
	if clientID == *property.AgentID {
		return nil, ErrVisitForbidden
	}

	// Prevent duplicate requests (same property + client + date).
	exists, err := s.visits.ExistsForSchedule(propertyID, clientID, req.ScheduledAt)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateVisit
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

	// Transition rules: only the agent confirms/completes; either party may
	// cancel; nobody manually sets a visit back to REQUESTED.
	isAgent := userID == visit.AgentID
	switch status {
	case models.VisitStatusConfirmed, models.VisitStatusCompleted:
		if !isAgent {
			return nil, ErrVisitForbidden
		}
	case models.VisitStatusCancelled:
		// allowed for both participants
	default:
		return nil, ErrVisitForbidden
	}

	if err := s.visits.UpdateStatus(visitID, status); err != nil {
		return nil, err
	}
	return s.visits.FindByID(visitID)
}
