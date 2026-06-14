package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type AlertHandler struct {
	svc *services.AlertService
}

func NewAlertHandler(svc *services.AlertService) *AlertHandler {
	return &AlertHandler{svc: svc}
}

// @Summary      Create a search alert
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateAlertRequest  true  "Search criteria"
// @Success      201   {object}  models.Alert
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /alerts [post]
func (h *AlertHandler) Create(c *gin.Context) {
	var req dto.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := middleware.CurrentUserID(c)
	alert, err := h.svc.Create(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmptyAlert), errors.Is(err, services.ErrInvalidCategory):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuplicateAlert):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create alert"})
		}
		return
	}
	c.JSON(http.StatusCreated, alert)
}

// @Summary      List my search alerts
// @Tags         alerts
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /alerts [get]
func (h *AlertHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch alerts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Delete a search alert
// @Tags         alerts
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Alert id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /alerts/{id} [delete]
func (h *AlertHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	userID := middleware.CurrentUserID(c)
	if err := h.svc.Delete(userID, id); err != nil {
		switch {
		case errors.Is(err, services.ErrAlertNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		case errors.Is(err, services.ErrNotAlertOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "this alert does not belong to you"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete alert"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
