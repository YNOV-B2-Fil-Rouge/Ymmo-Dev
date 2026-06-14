package services

import (
	"errors"
	"math"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrPropertyNotFound = errors.New("property not found")
	ErrNotPendingReview = errors.New("property is not awaiting review")
)

type PropertyService struct {
	properties *repositories.PropertyRepository
	users      *repositories.UserRepository
}

func NewPropertyService(properties *repositories.PropertyRepository, users *repositories.UserRepository) *PropertyService {
	return &PropertyService{properties: properties, users: users}
}

// Search returns a page of properties plus its pagination metadata.
func (s *PropertyService) Search(q dto.PropertySearchQuery) ([]models.Property, dto.Pagination, error) {
	items, total, err := s.properties.Search(q)
	if err != nil {
		return nil, dto.Pagination{}, err
	}

	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 12
	}
	if size > 50 {
		size = 50
	}

	meta := dto.Pagination{
		Page:       page,
		PageSize:   size,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(size))),
	}
	return items, meta, nil
}

// ListMine returns the agent's own properties (all statuses).
func (s *PropertyService) ListMine(agentID uint) ([]models.Property, error) {
	return s.properties.ListByAgent(agentID)
}

// ListManaged returns the properties a manager may oversee: HQ sees every
// property nationwide, a director only those of their own agency.
func (s *PropertyService) ListManaged(userID uint, role string) ([]models.Property, error) {
	if role == "HQ" {
		return s.properties.ListAll()
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.AgencyID == nil {
		return []models.Property{}, nil
	}
	return s.properties.ListByAgency(*user.AgencyID)
}

// Get returns a property and records a view (popularity tracking).
func (s *PropertyService) Get(id uint) (*models.Property, error) {
	property, err := s.properties.FindByID(id)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, ErrPropertyNotFound
	}
	// Best-effort: a failed counter update must not break the read.
	_ = s.properties.IncrementViewCount(id)
	return property, nil
}

// Create lists a new property. The workflow depends on who creates it:
//   - a SELLER (client) submits a listing that must be reviewed: it starts as
//     PENDING_REVIEW with no assigned agent (an agent validates it later);
//   - a staff member (AGENT/DIRECTOR/HQ) creates a DRAFT they own and publish.
func (s *PropertyService) Create(req dto.CreatePropertyRequest, creatorID uint, creatorRole string) (*models.Property, error) {
	property := &models.Property{
		Title:       req.Title,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Area:        req.Area,
		Rooms:       req.Rooms,
		Bedrooms:    req.Bedrooms,
		Bathrooms:   req.Bathrooms,
		Floor:       req.Floor,
		BuildYear:   req.BuildYear,
		City:        req.City,
		PostalCode:  req.PostalCode,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsExclusive: req.IsExclusive,
		AgencyID:    req.AgencyID,
	}

	if creatorRole == "SELLER" {
		// Submitted for validation; no agent owns it yet.
		property.Status = models.PropertyStatusPendingReview
		property.AgentID = nil
	} else {
		// Staff own their draft.
		property.Status = models.PropertyStatusDraft
		id := creatorID
		property.AgentID = &id
	}

	if req.Description != "" {
		property.Description = &req.Description
	}
	if req.EnergyRating != "" {
		property.EnergyRating = &req.EnergyRating
	}
	if req.GhgRating != "" {
		property.GhgRating = &req.GhgRating
	}
	if req.Address != "" {
		property.Address = &req.Address
	}

	if err := s.properties.Create(property); err != nil {
		return nil, err
	}
	return property, nil
}

// ListPending returns the seller submissions awaiting validation. HQ sees every
// agency; an agent/director only sees their own agency's submissions.
func (s *PropertyService) ListPending(userID uint, role string) ([]models.Property, error) {
	if role == "HQ" {
		return s.properties.ListPending(nil)
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.AgencyID == nil {
		return []models.Property{}, nil
	}
	return s.properties.ListPending(user.AgencyID)
}

// Validate approves a seller submission: the reviewing agent becomes the
// assigned agent and the property goes live (AVAILABLE).
func (s *PropertyService) Validate(propertyID, agentID uint) (*models.Property, error) {
	existing, err := s.properties.FindByID(propertyID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrPropertyNotFound
	}
	if existing.Status != models.PropertyStatusPendingReview {
		return nil, ErrNotPendingReview
	}
	updates := map[string]interface{}{
		"status":   models.PropertyStatusAvailable,
		"agent_id": agentID,
	}
	if err := s.properties.Update(propertyID, updates); err != nil {
		return nil, err
	}
	return s.properties.FindByID(propertyID)
}

// Update applies a partial change after checking the property exists.
func (s *PropertyService) Update(id uint, req dto.UpdatePropertyRequest) (*models.Property, error) {
	existing, err := s.properties.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrPropertyNotFound
	}

	updates := req.ToUpdates()
	if len(updates) > 0 {
		if err := s.properties.Update(id, updates); err != nil {
			return nil, err
		}
	}
	return s.properties.FindByID(id)
}

// Delete removes a property after checking it exists.
func (s *PropertyService) Delete(id uint) error {
	existing, err := s.properties.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrPropertyNotFound
	}
	return s.properties.Delete(id)
}
