package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/services"
)

type PhotoHandler struct {
	svc *services.PhotoService
}

func NewPhotoHandler(svc *services.PhotoService) *PhotoHandler {
	return &PhotoHandler{svc: svc}
}

// Add attaches a photo to a property (staff only).
// @Summary      Add a photo to a property
// @Tags         photos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                  true  "Property id"
// @Param        body  body      dto.AddPhotoRequest  true  "Photo URL"
// @Success      201   {object}  models.PropertyPhoto
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /properties/{id}/photos [post]
func (h *PhotoHandler) Add(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.AddPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	photo, err := h.svc.Add(id, req)
	if err != nil {
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not add photo"})
		return
	}
	c.JSON(http.StatusCreated, photo)
}

// Delete removes a photo (staff only).
// @Summary      Delete a property photo
// @Tags         photos
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int  true  "Property id"
// @Param        photoId  path  int  true  "Photo id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /properties/{id}/photos/{photoId} [delete]
func (h *PhotoHandler) Delete(c *gin.Context) {
	raw := c.Param("photoId")
	photoID, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || photoID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo id"})
		return
	}

	if err := h.svc.Delete(uint(photoID)); err != nil {
		if errors.Is(err, services.ErrPhotoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete photo"})
		return
	}
	c.Status(http.StatusNoContent)
}
