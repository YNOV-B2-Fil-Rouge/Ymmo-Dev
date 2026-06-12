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

// ListCollaborators returns the internal staff directory.
func (s *UserService) ListCollaborators() ([]models.User, error) {
	return s.users.ListInternal()
}
