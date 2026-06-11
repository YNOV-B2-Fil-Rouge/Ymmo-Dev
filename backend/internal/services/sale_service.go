package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrSaleNotFound       = errors.New("sale file not found")
	ErrNotSaleParticipant = errors.New("you are not involved in this sale file")
	ErrNotSaleAgent       = errors.New("only the sale's agent can do this")
	ErrBuyerNotFound      = errors.New("buyer not found")
)

type SaleService struct {
	sales      *repositories.SaleRepository
	properties *repositories.PropertyRepository
	users      *repositories.UserRepository
}

func NewSaleService(sales *repositories.SaleRepository, properties *repositories.PropertyRepository, users *repositories.UserRepository) *SaleService {
	return &SaleService{sales: sales, properties: properties, users: users}
}

// Create opens a sale file. The current user (an agent) handles it.
func (s *SaleService) Create(agentID uint, req dto.CreateSaleRequest) (*models.SaleFile, error) {
	property, err := s.properties.FindByID(req.PropertyID)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, ErrPropertyNotFound
	}

	buyer, err := s.users.FindByID(req.BuyerID)
	if err != nil {
		return nil, err
	}
	if buyer == nil {
		return nil, ErrBuyerNotFound
	}

	sale := &models.SaleFile{
		PropertyID:      req.PropertyID,
		BuyerID:         req.BuyerID,
		AgentID:         agentID,
		Status:          models.SaleStatusOffer,
		NegotiatedPrice: req.NegotiatedPrice,
		OfferDate:       req.OfferDate,
	}
	if err := s.sales.Create(sale); err != nil {
		return nil, err
	}
	return sale, nil
}

// List returns the user's sale files (as buyer or agent).
func (s *SaleService) List(userID uint) ([]models.SaleFile, error) {
	return s.sales.ListForUser(userID)
}

// Get returns a sale file if the user is a participant.
func (s *SaleService) Get(userID, saleID uint) (*models.SaleFile, error) {
	sale, err := s.sales.FindByID(saleID)
	if err != nil {
		return nil, err
	}
	if sale == nil {
		return nil, ErrSaleNotFound
	}
	if sale.BuyerID != userID && sale.AgentID != userID {
		return nil, ErrNotSaleParticipant
	}
	return sale, nil
}

// Update advances a sale file (the sale's agent only). When the sale is marked
// COMPLETED, the linked property is automatically set to SOLD.
func (s *SaleService) Update(userID, saleID uint, req dto.UpdateSaleRequest) (*models.SaleFile, error) {
	sale, err := s.sales.FindByID(saleID)
	if err != nil {
		return nil, err
	}
	if sale == nil {
		return nil, ErrSaleNotFound
	}
	if sale.AgentID != userID {
		return nil, ErrNotSaleAgent
	}

	updates := req.ToUpdates()
	if len(updates) > 0 {
		if err := s.sales.Update(saleID, updates); err != nil {
			return nil, err
		}
	}

	// Business rule: a completed sale marks its property as sold.
	if req.Status != nil && *req.Status == models.SaleStatusCompleted {
		if err := s.properties.Update(sale.PropertyID, map[string]interface{}{"status": models.PropertyStatusSold}); err != nil {
			return nil, err
		}
	}

	return s.sales.FindByID(saleID)
}
