package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type MeetingHandler struct {
	svc *services.MeetingService
}

func NewMeetingHandler(svc *services.MeetingService) *MeetingHandler {
	return &MeetingHandler{svc: svc}
}

// Create schedules an internal meeting (staff only).
// @Summary      Create a meeting
// @Tags         meetings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateMeetingRequest  true  "Meeting details"
// @Success      201   {object}  models.Meeting
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /meetings [post]
func (h *MeetingHandler) Create(c *gin.Context) {
	var req dto.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	organizerID := middleware.CurrentUserID(c)
	meeting, err := h.svc.Create(organizerID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidMeetingTime):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuplicateMeeting):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create meeting"})
		}
		return
	}
	c.JSON(http.StatusCreated, meeting)
}

// List returns the current user's meetings.
// @Summary      List my meetings
// @Tags         meetings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /meetings [get]
func (h *MeetingHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch meetings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Delete cancels a meeting (organizer only).
// @Summary      Delete a meeting
// @Tags         meetings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Meeting id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /meetings/{id} [delete]
func (h *MeetingHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	userID := middleware.CurrentUserID(c)
	if err := h.svc.Delete(userID, id); err != nil {
		switch {
		case errors.Is(err, services.ErrMeetingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "meeting not found"})
		case errors.Is(err, services.ErrNotOrganizer):
			c.JSON(http.StatusForbidden, gin.H{"error": "only the organizer can delete this meeting"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete meeting"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
