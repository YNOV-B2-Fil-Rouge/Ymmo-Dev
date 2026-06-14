package services

import (
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(users *repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) ListAllUsers() ([]models.User, error) {
	return s.users.ListAll()
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
