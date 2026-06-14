// Package router assembles the Gin engine: middleware, wiring and routes.
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"ymmo/internal/config"
	"ymmo/internal/handlers"
	"ymmo/internal/middleware"
	"ymmo/internal/repositories"
	"ymmo/internal/services"
	"ymmo/internal/validation"

	_ "ymmo/docs"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	validation.Register()

	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	health := handlers.NewHealthHandler(db)
	r.GET("/health", health.Check)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Static("/uploads", cfg.UploadDir)

	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	authService := services.NewAuthService(userRepo, roleRepo, cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authService)

	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	permissionRepo := repositories.NewPermissionRepository(db)
	permissionService := services.NewPermissionService(permissionRepo)
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	propertyRepo := repositories.NewPropertyRepository(db)
	propertyService := services.NewPropertyService(propertyRepo, userRepo)
	propertyHandler := handlers.NewPropertyHandler(propertyService)

	favoriteRepo := repositories.NewFavoriteRepository(db)
	favoriteService := services.NewFavoriteService(favoriteRepo, propertyRepo)
	favoriteHandler := handlers.NewFavoriteHandler(favoriteService)

	photoRepo := repositories.NewPhotoRepository(db)
	photoService := services.NewPhotoService(photoRepo, propertyRepo)
	photoHandler := handlers.NewPhotoHandler(photoService, cfg.UploadDir, cfg.PublicURL)

	conversationRepo := repositories.NewConversationRepository(db)
	messageService := services.NewMessageService(conversationRepo, userRepo)
	messageHandler := handlers.NewMessageHandler(messageService)

	visitRepo := repositories.NewVisitRepository(db)
	visitService := services.NewVisitService(visitRepo, propertyRepo)
	visitHandler := handlers.NewVisitHandler(visitService)

	meetingRepo := repositories.NewMeetingRepository(db)
	meetingService := services.NewMeetingService(meetingRepo)
	meetingHandler := handlers.NewMeetingHandler(meetingService)

	alertRepo := repositories.NewAlertRepository(db)
	alertService := services.NewAlertService(alertRepo)
	alertHandler := handlers.NewAlertHandler(alertService)

	saleRepo := repositories.NewSaleRepository(db)
	saleService := services.NewSaleService(saleRepo, propertyRepo, userRepo)
	saleHandler := handlers.NewSaleHandler(saleService)

	sellerAppRepo := repositories.NewSellerApplicationRepository(db)
	sellerAppService := services.NewSellerApplicationService(sellerAppRepo, userRepo, roleRepo)
	sellerAppHandler := handlers.NewSellerApplicationHandler(sellerAppService)

	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.Auth(cfg.JWTSecret), authHandler.Me)
		}

		properties := api.Group("/properties")
		{
			properties.GET("", propertyHandler.List)
			properties.GET("/:id", propertyHandler.Get)

			properties.POST("/:id/favori