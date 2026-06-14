package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type SaleHandler struct {
	svc *services.SaleService
}

func NewSaleHandler(svc *services.SaleService) *SaleHandler {
	return &SaleHandler{svc: svc}
}

// Create opens a sale file (staff only).
// @Summary      Open a sale file
// @Tags         sales
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateSaleRequest  true  "Property and buyer"
// @Success      201   {object}  models.SaleFile
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /sales [post]
func (h *SaleHandler) Create(c *gin.Context) {
	var req dto.CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	agentID := middleware.CurrentUserID(c)
	sale, err := h.svc.Create(agentID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPropertyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
		case errors.Is(err, services.ErrBuyerNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "buyer not found"})
		case errors.Is(err, services.ErrNotABuyer), errors.Is(err, services.ErrPastOffer):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuplicateSale):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not open sale file"})
		}
		return
	}
	c.JSON(http.StatusCreated, sale)
}

// List returns the current user's sale files.
// @Summary      List my sale files
// @Tags         sales
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /sales [get]
func (h *SaleHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch sale files"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Get returns one sale file (participant only).
// @Summary      Get a sale file
// @Tags         sales
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Sale file id"
// @Success      200  {object}  models.SaleFile
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /sales/{id} [get]
func (h *SaleHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	userID := middleware.CurrentUserID(c)
	sale, err := h.svc.Get(userID, id)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}
	c.JSON(http.StatusOK, sale)
}

// Update advances a sale file (the sale's agent only).
// @Summary      Update a sale file
// @Tags         sales
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                   true  "Sale file id"
// @Param        body  body      dto.UpdateSaleRequest  true  "Fields to update"
// @Success      200   {object}  models.SaleFile
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /sales/{id} [patch]
func (h *SaleHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := middleware.CurrentUserID(c)
	sale, err := h.svc.Update(userID, id, req)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}
	c.JSON(http.StatusOK, sale)
}

// writeAccessError maps sale errors to HTTP status codes.
func (h *SaleHandler) writeAccessError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrSaleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "sale file not found"})
	case errors.Is(err, services.ErrNotSaleParticipant):
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not involved in this sale file"})
	case errors.Is(err, services.ErrNotSaleAgent):
		c.JSON(http.StatusForbidden, gin.H{"error": "only the sale's agent can do this"})
	case errors.Is(err, services.ErrInvalidSaleTransition):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale status transition (steps must be sequential)"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process sale file"})
	}
}
