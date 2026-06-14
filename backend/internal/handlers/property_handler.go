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

// @Summary      List / search properties
// @Description  Public catalogue. Only AVAILABLE properties are returned by default.
// @Tags         properties
// @Produce      json
// @Param        city         query     string  false  "City"
// @Param        category_id  query     int     false  "Category id"
// @Param        sector       query     string  false  "RESIDENTIAL or COMMERCIAL"
// @Param        min_price    query     number  false  "Minimum price"
// @Param        max_price    query     number  false  "Maximum price"
// @Param        min_area     query     number  false  "Minimum area (m2)"
// @Param        max_area     query     number  false  "Maximum area (m2)"
// @Param        max_energy   query     string  false  "Max energy class (A..G, returns this or better)"
// @Param        page         query     int     false  "Page number (default 1)"
// @Param        page_size    query     int     false  "Page size (default 12, max 50)"
// @Param        sort         query     string  false  "price_asc | price_desc | recent"
// @Success      200          {object}  map[string]interface{}
// @Router       /properties [get]
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

// @Summary      List my properties
// @Tags         properties
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /me/properties [get]
func (h *PropertyHandler) ListMine(c *gin.Context) {
	agentID := middleware.CurrentUserID(c)
	items, err := h.svc.ListMine(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch your properties"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      List all properties (management)
// @Tags         properties
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/properties [get]
func (h *PropertyHandler) ListAll(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	items, err := h.svc.ListManaged(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch properties"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Get a property by id
// @Description  Returns the property and increments its view counter.
// @Tags         properties
// @Produce      json
// @Param        id   path      int  true  "Property id"
// @Success      200  {object}  models.Property
// @Failure      404  {object}  map[string]string
// @Router       /properties/{id} [get]
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

// @Summary      Create a property
// @Description  Staff create a DRAFT they own; a seller submits a PENDING_REVIEW listing for validation.
// @Tags         properties
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreatePropertyRequest  true  "Property to create"
// @Success      201   {object}  models.Property
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /properties [post]
func (h *PropertyHandler) Create(c *gin.Context) {
	var req dto.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	creatorID := middleware.CurrentUserID(c)
	creatorRole := middleware.CurrentRole(c)
	property, err := h.svc.Create(req, creatorID, creatorRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create property"})
		return
	}
	c.JSON(http.StatusCreated, property)
}

// @Summary      List properties awaiting validation
// @Description  Seller submissions in PENDING_REVIEW. Scoped to the staff member's agency (HQ sees all).
// @Tags         properties
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/pending-properties [get]
func (h *PropertyHandler) ListPending(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	items, err := h.svc.ListPending(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch pending properties"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Validate a seller submission
// @Tags         properties
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Property id"
// @Success      200  {object}  models.Property
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /properties/{id}/validate [post]
func (h *PropertyHandler) Validate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	agentID := middleware.CurrentUserID(c)
	property, err := h.svc.Validate(id, agentID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPropertyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
		case errors.Is(err, services.ErrNotPendingReview):
			c.JSON(http.StatusBadRequest, gin.H{"error": "this property is not awaiting validation"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate property"})
		}
		return
	}
	c.JSON(http.StatusOK, property)
}

// @Summary      Update a property
// @Description  Staff only. Partial update: only the fields sent are changed.
// @Tags         properties
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                        true  "Property id"
// @Param        body  body      dto.UpdatePropertyRequest  true  "Fields to update"
// @Success      200   {object}  models.Property
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /properties/{id} [put]
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

// @Summary      Delete a property
// @Description  Staff only.
// @Tags         properties
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Property id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /properties/{id} [delete]
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

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}
