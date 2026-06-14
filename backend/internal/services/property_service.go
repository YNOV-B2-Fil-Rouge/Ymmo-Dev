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

func (s *PropertyService) ListMine(agentID uint) ([]models.Property, error) {
	return s.properties.ListByAgent(agentID)
}

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

func (s *PropertyService) Get(id uint) (*models.Property, error) {
	property, err := s.properties.FindByID(id)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, ErrPropertyNotFound
	}
	_ = s.properties.IncrementViewCount(id)
	return property, nil
}

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
		property.Status = models.PropertyStatusPendingReview
		property.AgentID = nil
	} else {