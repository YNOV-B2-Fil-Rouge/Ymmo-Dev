package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type FavoriteHandler struct {
	svc *services.FavoriteService
}

func NewFavoriteHandler(svc *services.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

// @Summary      Add a property to favorites
// @Tags         favorites
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Property id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /properties/{id}/favorites [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.svc.Add(userID, id); err != nil {
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		if errors.Is(err, services.ErrPropertyNotPublic) {
			c.JSON(http.StatusConflict, gin.H{"error": "property is not available"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not add favorite"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Remove a property from favorites
// @Tags         favorites
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Property id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Router       /properties/{id}/favorites [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.svc.Remove(userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not remove favorite"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      List my favorite properties
// @Tags         favorites
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch favorites"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
