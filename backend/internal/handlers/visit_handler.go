package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type VisitHandler struct {
	svc *services.VisitService
}

func NewVisitHandler(svc *services.VisitService) *VisitHandler {
	return &VisitHandler{svc: svc}
}

// Request asks to visit a property.
// @Summary      Request a property visit
// @Tags         visits
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                      true  "Property id"
// @Param        body  body      dto.RequestVisitRequest  true  "Schedule and notes"
// @Success      201   {object}  models.Visit
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /properties/{id}/visits [post]
func (h *VisitHandler) Request(c *gin.Context) {
	propertyID, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.RequestVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	clientID := middleware.CurrentUserID(c)
	clientRole := middleware.CurrentRole(c)
	visit, err := h.svc.Request(clientID, clientRole, propertyID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPropertyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
		case errors.Is(err, services.ErrVisitForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to request this visit"})
		case errors.Is(err, services.ErrPropertyNotAvailable), errors.Is(err, services.ErrDuplicateVisit):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrPastSchedule), errors.Is(err, services.ErrNoAgentForProperty):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not request visit"})
		}
		return
	}
	c.JSON(http.StatusCreated, visit)
}

// List returns the current user's visits.
// @Summary      List my visits
// @Tags         visits
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /visits [get]
func (h *VisitHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch visits"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// UpdateStatus changes a visit's status (confirm / cancel / complete).
// @Summary      Update a visit status
// @Tags         visits
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                            true  "Visit id"
// @Param        body  body      dto.UpdateVisitStatusRequest   true  "New status"
// @Success      200   {object}  models.Visit
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /visits/{id} [patch]
func (h *VisitHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateVisitStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := middleware.CurrentUserID(c)
	visit, err := h.svc.UpdateStatus(userID, id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrVisitNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "visit not found"})
		case errors.Is(err, services.ErrNotVisitParticipant):
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not a participant of this visit"})
		case errors.Is(err, services.ErrVisitForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "only the agent can confirm or complete a visit"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update visit"})
		}
		return
	}
	c.JSON(http.StatusOK, visit)
}
