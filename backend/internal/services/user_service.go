package services

import (
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

// UserService exposes read operations on users for management views.
type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(users *repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

// ListAllUsers returns every user (IT user management).
func (s *UserService) ListAllUsers() ([]models.User, error) {
	return s.users.ListAll()
}

// ListCollaborators returns the staff directory: nationwide for HQ, agency-
// scoped for a director.
func (s *UserService) ListCollaborators(userID uint, role string) ([]models.User, error) {
	if role == "HQ" {
		return s.users.ListInternal()
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.AgencyID == nil {
		return []models.User{}, nil
	}
	return s.users.ListInternalByAgency(*user.AgencyID)
}
