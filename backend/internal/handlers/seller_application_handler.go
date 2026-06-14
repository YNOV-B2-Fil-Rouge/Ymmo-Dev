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

type SellerApplicationHandler struct {
	svc *services.SellerApplicationService
}

func NewSellerApplicationHandler(svc *services.SellerApplicationService) *SellerApplicationHandler {
	return &SellerApplicationHandler{svc: svc}
}

// Apply lets a buyer request to become a seller.
// @Summary      Apply to become a seller
// @Description  Buyer only. Creates a PENDING application reviewed by an agent.
// @Tags         seller-applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.ApplyAsSellerRequest  false  "Optional motivation"
// @Success      201   {object}  models.SellerApplication
// @Failure      403   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /seller-applications [post]
func (h *SellerApplicationHandler) Apply(c *gin.Context) {
	var req dto.ApplyAsSellerRequest
	// Body is optional; ignore "EOF" when nothing is sent.
	_ = c.ShouldBindJSON(&req)

	userID := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	app, err := h.svc.Apply(userID, role, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotBuyerRole):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationPending):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit application"})
		}
		return
	}
	c.JSON(http.StatusCreated, app)
}

// Mine returns the current user's latest application (or null).
// @Summary      My seller application
// @Tags         seller-applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /seller-applications [get]
func (h *SellerApplicationHandler) Mine(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	app, err := h.svc.Mine(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch your application"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": app})
}

// ListPending returns applications awaiting review (staff only).
// @Summary      List pending seller applications
// @Tags         seller-applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /management/seller-applications [get]
func (h *SellerApplicationHandler) ListPending(c *gin.Context) {
	items, err := h.svc.ListPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch applications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Approve validates an application and promotes the user to seller (staff only).
// @Summary      Approve a seller application
// @Tags         seller-applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Application id"
// @Success      200  {object}  models.SellerApplication
// @Router       /management/seller-applications/{id}/approve [post]
func (h *SellerApplicationHandler) Approve(c *gin.Context) {
	h.review(c, true)
}

// Reject declines an application (staff only).
// @Summary      Reject a seller application
// @Tags         seller-applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Application id"
// @Success      200  {object}  models.SellerApplication
// @Router       /management/seller-applications/{id}/reject [post]
func (h *SellerApplicationHandler) Reject(c *gin.Context) {
	h.review(c, false)
}

func (h *SellerApplicationHandler) review(c *gin.Context, approve bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	reviewerID := middleware.CurrentUserID(c)

	var app interface{}
	var aErr error
	if approve {
		app, aErr = h.svc.Approve(id, reviewerID)
	} else {
		app, aErr = h.svc.Reject(id, reviewerID)
	}

	if aErr != nil {
		switch {
		case errors.Is(aErr, services.ErrApplicationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": aErr.Error()})
		case errors.Is(aErr, services.ErrApplicationReviewed):
			c.JSON(http.StatusBadRequest, gin.H{"error": aErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process application"})
		}
		return
	}
	c.JSON(http.StatusOK, app)
}
