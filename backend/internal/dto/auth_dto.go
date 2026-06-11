// Package dto holds the validated request/response shapes exchanged over
// HTTP. Keeping them separate from the GORM models means we never leak
// internal fields (like password_hash) and we validate every input.
package dto

import "ymmo/internal/models"

// RegisterRequest is the public sign-up payload.
// `binding` tags are validated automatically by Gin (go-playground/validator).
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8,max=72"` // bcrypt caps at 72 bytes
	FirstName string `json:"first_name" binding:"required,min=2,max=80"`
	LastName  string `json:"last_name" binding:"required,min=2,max=80"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
}

// LoginRequest is the credentials payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse is the safe, public view of a user.
type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// AuthResponse is returned on successful login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// NewUserResponse maps an internal model to its safe public representation.
func NewUserResponse(u *models.User) UserResponse {
	role := ""
	if u.Role != nil {
		role = u.Role.Code
	}
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      role,
	}
}
