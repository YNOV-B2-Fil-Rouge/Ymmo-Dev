package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrNotBuyerRole        = errors.New("only a buyer can apply to become a seller")
	ErrApplicationPending  = errors.New("you already have a pending application")
	ErrApplicationNotFound = errors.New("seller application not found")
	ErrApplicationReviewed = errors.New("this application has already been reviewed")
	ErrSellerRoleMissing   = errors.New("SELLER role not found (check seed data)")
)

// SellerApplicationService implements the buyer→seller promotion workflow.
type SellerApplicationService struct {
	apps  *repositories.SellerApplicationRepository
	users *repositories.UserRepository
	roles *repositories.RoleRepository
}

func NewSellerApplicationService(
	apps *repositories.SellerApplicationRepository,
	users *repositories.UserRepository,
	roles *repositories.RoleRepository,
) *SellerApplicationService {
	return &SellerApplicationService{apps: apps, users: users, roles: roles}
}

// Apply records a buyer's request. Only a BUYER may apply, and only once at a
// time (no duplicate pending request).
func (s *SellerApplicationService) Apply(userID uint, role string, req dto.ApplyAsSellerRequest) (*models.SellerApplication, error) {
	if role != "BUYER" {
		return nil, ErrNotBuyerRole
	}
	pending, err := s.apps.HasPending(userID)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrApplicationPending
	}

	app := &models.SellerApplication{
		UserID: userID,
		Status: models.SellerApplicationPending,
	}
	if req.Motivation != "" {
		app.Motivation = &req.Motivation
	}
	if err := s.apps.Create(app); err != nil {
		return nil, err
	}
	return app, nil
}

// Mine returns the user's most recent application (or nil if they never applied).
func (s *SellerApplicationService) Mine(userID uint) (*models.SellerApplication, error) {
	return s.apps.FindLatestByUser(userID)
}

// ListPending returns the applications awaiting an agent's review.
func (s *SellerApplicationService) ListPending() ([]models.SellerApplication, error) {
	return s.apps.ListPending()
}

// Approve validates an application AND promotes the applicant to SELLER.
func (s *SellerApplicationService) Approve(appID, reviewerID uint) (*models.SellerApplication, error) {
	app, err := s.loadReviewable(appID)
	if err != nil {
		return nil, err
	}

	sellerRole, err := s.roles.FindByCode("SELLER")
	if err != nil {
		return nil, err
	}
	if sellerRole == nil {
		return nil, ErrSellerRoleMissing
	}
	if err := s.users.UpdateRole(app.UserID, sellerRole.ID); err != nil {
		return nil, err
	}
	if err := s.apps.Review(appID, models.SellerApplicationApproved, reviewerID); err != nil {
		return nil, err
	}
	return s.apps.FindByID(appID)
}

// Reject declines an application without changing the user's role.
func (s *SellerApplicationService) Reject(appID, reviewerID uint) (*models.SellerApplication, error) {
	if _, err := s.loadReviewable(appID); err != nil {
		return nil, err
	}
	if err := s.apps.Review(appID, models.SellerApplicationRejected, reviewerID); err != nil {
		return nil, err
	}
	return s.apps.FindByID(appID)
}

// loadReviewable fetches an application and ensures it is still pending.
func (s *SellerApplicationService) loadReviewable(appID uint) (*models.SellerApplication, error) {
	app, err := s.apps.FindByID(appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrApplicationNotFound
	}
	if app.Status != models.SellerApplicationPending {
		return nil, ErrApplicationReviewed
	}
	return app, nil
}
