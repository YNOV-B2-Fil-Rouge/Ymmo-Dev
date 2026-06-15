// Package dto holds the validated request/response shapes.
package dto

import "ymmo/internal/models"

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	FirstName string `json:"first_name" binding:"required,min=2,max=80"`
	LastName  string `json:"last_name" binding:"required,min=2,max=80"`
	Phone     string `json:"phone" binding:"omitempty,phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

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
