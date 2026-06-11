// Package router assembles the Gin engine: global middleware + routes.
// Adding a new module later = registering its routes here, nothing else.
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ymmo/internal/config"
	"ymmo/internal/handlers"
)

// New builds the fully configured HTTP engine.
func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware: structured logging + panic recovery.
	r.Use(gin.Logger(), gin.Recovery())

	// CORS so the Tailwind front-end (served from another origin during
	// development) can call the API.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- Health check (outside the versioned API) ---
	health := handlers.NewHealthHandler(db)
	r.GET("/health", health.Check)

	// --- Versioned API ---
	// Future modules (auth, properties, ...) register their routes here.
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

	return r
}
