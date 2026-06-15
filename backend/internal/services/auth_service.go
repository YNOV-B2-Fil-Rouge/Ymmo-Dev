// Package services holds the business logic, orchestrating the repositories.
package services

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
	"ymmo/internal/security"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveAccount    = errors.New("account is inactive")
	ErrDefaultRoleMissing = errors.New("default BUYER role not found (check seed data)")
)

const tokenTTL = 24 * time.Hour

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("ymmo-timing-guard"), bcrypt.DefaultCost)

type AuthService struct {
	users     *repositories.UserRepository
	roles     *repositories.RoleRepository
	jwtSecret string
}

func NewAuthService(users *repositories.UserRepository, roles *repositories.RoleRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, roles: roles, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*models.User, error) {
	taken, err := s.users.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrEmailTaken
	}

	role, err := s.roles.FindByCode("BUYER")
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrDefaultRoleMissing
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleID:       role.ID,
		IsActive:     true,
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	user.Role = role
	return user, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (string, *models.User, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		return "", nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return "", nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return "", nil, ErrInactiveAccount
	}

	roleCode := ""
	if user.Role != nil {
		roleCode = user.Role.Code
	}
	token, err := security.GenerateToken(s.jwtSecret, user.ID, roleCode, tokenTTL)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) GetUser(id uint) (*models.User, error) {
	return s.users.FindByID(id)
}
