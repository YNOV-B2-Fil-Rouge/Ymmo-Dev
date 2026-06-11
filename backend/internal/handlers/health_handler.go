// Package handlers contains the HTTP controllers (the "C" of MVC).
// Each handler is thin: it parses the request, calls a service, and
// writes a JSON response. Business logic lives in the service layer.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler exposes liveness/readiness checks.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler builds the handler with its dependencies injected
// (Dependency Inversion: it depends on *gorm.DB passed in, not on a
// global variable).
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check reports whether the API and its database are reachable.
// GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "up"
	httpStatus := http.StatusOK

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "down"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":   "ok",
		"service":  "ymmo-api",
		"database": dbStatus,
	})
}
