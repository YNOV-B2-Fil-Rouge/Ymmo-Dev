package services

import (
	"errors"

	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(users *repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) ListAllUsers() ([]models.User, error) {
	return s.users.ListAll()
}

// DeleteAccount soft-deletes a user (anonymize + deactivate).
func (s *UserService) DeleteAccount(id uint) error {
	user, err := s.users.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	return s.users.SoftDelete(id)
}

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
