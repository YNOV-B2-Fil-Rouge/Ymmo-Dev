package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// @Summary      List all users (IT)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/users [get]
func (h *UserHandler) ListAll(c *gin.Context) {
	items, err := h.svc.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      List collaborators
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/collaborators [get]
func (h *UserHandler) ListCollaborators(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	items, err := h.svc.ListCollaborators(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch collaborators"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
