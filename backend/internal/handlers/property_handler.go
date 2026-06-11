package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type PropertyHandler struct {
	svc *services.PropertyService
}

func NewPropertyHandler(svc *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{svc: svc}
}

// List returns the public, filtered, paginated catalogue.
// GET /api/v1/properties
func (h *PropertyHandler) List(c *gin.Context) {
	var q dto.PropertySearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filters", "details": err.Error()})
		return
	}

	items, pagination, err := h.svc.Search(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch properties"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"pagination": pagination,
	})
}

// Get returns a single property and records a view.
// GET /api/v1/properties/:id
func (h *PropertyHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	property, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch property"})
		return
	}
	c.JSON(http.StatusOK, property)
}

// Create lists a new property (agent/director/HQ only).
// POST /api/v1/properties
func (h *PropertyHandler) Create(c *gin.Context) {
	var req dto.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	agentID := middleware.CurrentUserID(c)
	property, err := h.svc.Create(req, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create property"})
		return
	}
	c.JSON(http.StatusCreated, property)
}

// Update applies a partial change to a property.
// PUT /api/v1/properties/:id
func (h *PropertyHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	property, err := h.svc.Update(id, req)
	if err != nil {
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update property"})
		return
	}
	c.JSON(http.StatusOK, property)
}

// Delete removes a property.
// DELETE /api/v1/properties/:id
func (h *PropertyHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete property"})
		return
	}
	c.Status(http.StatusNoContent)
}

// parseID reads and validates the :id path parameter. It writes a 400 and
// returns ok=false when the value is not a positive integer.
func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}
