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
	uploadDir  string // where files are written on disk
	publicBase string // public base URL used to build the photo URL
}

func NewPhotoHandler(svc *services.PhotoService, uploadDir, publicBase string) *PhotoHandler {
	return &PhotoHandler{svc: svc, uploadDir: uploadDir, publicBase: publicBase}
}

// allowedImageExt lists the image types we accept for upload.
var allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}

const maxPhotoSize = 5 << 20 // 5 MB

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

// Upload stores an uploaded image file on disk and registers it as a property
// photo (staff only). The file is saved under the upload directory (a Docker
// volume) and the DB keeps only its public URL.
// @Summary      Upload a photo file for a property
// @Tags         photos
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      int     true   "Property id"
// @Param        file        formData  file    true   "Image file (jpg/png/webp/gif, max 5MB)"
// @Param        is_primary  formData  bool    false  "Set as the primary photo"
// @Success      201   {object}  models.PropertyPhoto
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /properties/{id}/photos/upload [post]
func (h *PhotoHandler) Upload(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided (form field 'file')"})
		return
	}
	if fileHeader.Size > maxPhotoSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImageExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type (jpg, png, webp, gif)"})
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare upload directory"})
		return
	}

	// Unique, non-guessable filename: <unixnano>-<random>.<ext>.
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	name := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(buf), ext)
	dst := filepath.Join(h.uploadDir, name)

	if err := c.SaveUploadedFile(fileHeader, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}

	publicURL := strings.TrimRight(h.publicBase, "/") + "/uploads/" + name
	photo, err := h.svc.Add(id, dto.AddPhotoRequest{URL: publicURL, IsPrimary: c.PostForm("is_primary") == "true"})
	if err != nil {
		_ = os.Remove(dst) // roll back the file if the DB insert failed
		if errors.Is(err, services.ErrPropertyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save photo"})
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
