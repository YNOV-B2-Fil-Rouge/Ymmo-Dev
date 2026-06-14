package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/services"
)

type PhotoHandler struct {
	svc        *services.PhotoService
	uploadDir  string
	publicBase string
}

func NewPhotoHandler(svc *services.PhotoService, uploadDir, publicBase string) *PhotoHandler {
	return &PhotoHandler{svc: svc, uploadDir: uploadDir, publicBase: publicBase}
}

var allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}

const maxPhotoSize = 5 << 20

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

// @Summary      Upload a photo file for a property
// @Tags         photos
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      int     true   "Property id"
// @Param        file        formData  file    true   "Image file (jpg/png/webp/gif, max 5MB)"
// @Param        is_primary  formData  bool    false 