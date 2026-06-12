package services

import (
	"errors"
	"math"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var ErrPropertyNotFound = errors.New("property not found")

type PropertyService struct {
	properties *repositories.PropertyRepository
}

func NewPropertyService(properties *repositories.PropertyRepository) *PropertyService {
	return &PropertyService{properties: properties}
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

// Create lists a new property. It starts as a DRAFT: nothing is public until
// it is reviewed/published (workflow handled in a later module).
func (s *PropertyService) Create(req dto.CreatePropertyRequest, agentID uint) (*models.Property, error) {
	property := &models.Property{
		Title:       req.Title,
		CategoryID:  req.CategoryID,
		Status:      models.PropertyStatusDraft,
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
		AgentID:     &agentID,
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
