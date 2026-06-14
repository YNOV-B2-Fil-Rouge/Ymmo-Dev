package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/services"
)

type PermissionHandler struct {
	svc *services.PermissionService
}

func NewPermissionHandler(svc *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

// @Summary      Get the access-rights matrix
// @Tags         permissions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/permissions [get]
func (h *PermissionHandler) GetMatrix(c *gin.Context) {
	rows, err := h.svc.Matrix()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch permissions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}
