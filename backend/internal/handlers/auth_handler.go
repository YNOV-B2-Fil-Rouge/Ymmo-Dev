package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

// AuthHandler exposes the authentication endpoints.
type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register creates a new buyer account.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	user, err := h.svc.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create account"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.NewUserResponse(user))
}

// Login authenticates and returns a JWT.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	token, user, err := h.svc.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrInactiveAccount):
			// Same message for both -> no information leak about which failed.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  dto.NewUserResponse(user),
	})
}

// Me returns the currently authenticated user's profile.
// GET /api/v1/auth/me   (protected)
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	user, err := h.svc.GetUser(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, dto.NewUserResponse(user))
}
