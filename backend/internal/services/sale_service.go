package services

import (
	"errors"
	"time"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrSaleNotFound          = errors.New("sale file not found")
	ErrNotSaleParticipant    = errors.New("you are not involved in this sale file")
	ErrNotSaleAgent          = errors.New("only the sale's agent can do this")
	ErrBuyerNotFound         = errors.New("buyer not found")
	ErrNotABuyer             = errors.New("the buyer_id must reference a buyer")
	ErrPastOffer             = errors.New("the offer date cannot be in the past")
	ErrDuplicateSale         = errors.New("a sale file already exists for this property and buyer")
	ErrInvalidSaleTransition = errors.New("invalid sale status transition")
)

var saleTransitions = map[string][]string{
	models.SaleStatusOffer:               {models.SaleStatusPreliminaryContract, models.SaleStatusCancelled},
	models.SaleStatusPreliminaryContract: {models.SaleStatusDeed, models.SaleStatusCancelled},
	models.SaleStatusDeed:                {models.SaleStatusCompleted, models.SaleStatusCancelled},
}

func isValidSaleTransition(current, next string) bool {
	if current == next {
		return true
	}
	for _, allowed := range saleTransitions[current] {
		if allowed == next {
			return true
		}
	}
	return false
}

type SaleService struct {
	sales      *repositories.SaleRepository
	properties *repositories.PropertyRepository
	users      *repositories.UserRepository
}

func NewSaleService(sales *repositories.SaleRepository, properties *repositories.PropertyRepository, users *repositories.UserRepository) *SaleService {
	return &SaleService{sales: sales, properties: properties, users: users}
}

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
	if buyer.Role == nil || buyer.Role.Code != "BUYER" {
		return nil, ErrNotABuyer
	}

	if req.OfferDate != nil && req.OfferDate.Before(time.Now().Truncate(24*time.Hour)) {
		return nil, ErrPastOffer
	}

	exists, err := s.sales.ExistsActive(req.PropertyID, req.BuyerID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateSale
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

func (s *SaleService) List(userID uint) ([]models.SaleFile, error) {
	return s.sales.ListForUser(userID)
}

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

	if req.Status != nil && !isValidSaleTransition(sale.Status, *req.Status) {
		return nil, ErrInvalidSaleTransition
	}

	updates := req.ToUpdates()
	if len(updates) > 0 {
		if err := s.sales.Update(saleID, updates); err != nil {
			return nil, err
		}
	}

	if req.Status != nil && *req.Status == models.SaleStatusCompleted {
		if err := s.properties.Update(sale.PropertyID, map[string]interface{}{"status": models.PropertyStatusSold}); err != nil {
			return nil, err
		}
	}

	return s.sales.FindByID(saleID)
}
