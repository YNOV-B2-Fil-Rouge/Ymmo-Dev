package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/services"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ListCollaborators returns the internal staff directory (director/HQ only).
// @Summary      List collaborators
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/collaborators [get]
func (h *UserHandler) ListCollaborators(c *gin.Context) {
	items, err := h.svc.ListCollaborators()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch collaborators"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
