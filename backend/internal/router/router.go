// Package router assembles the Gin engine: global middleware + routes.
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
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	validation.Register()

	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery())

	// Auth uses Bearer tokens (no cookies), so credentials aren't needed and we can
	// allow any origin by default. With the nginx reverse proxy the front is anyway
	// same-origin; set CORS_ALLOWED_ORIGINS to restrict it in a stricter setup.
	corsCfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}
	if len(cfg.CORSOrigins) > 0 {
		corsCfg.AllowOrigins = cfg.CORSOrigins
	} else {
		corsCfg.AllowAllOrigins = true
	}
	r.Use(cors.New(corsCfg))

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
		properties.Use(middleware.OptionalAuth(cfg.JWTSecret))
		{
			properties.GET("", propertyHandler.List)
			properties.GET("/:id", propertyHandler.Get)

			properties.POST("/:id/favorites", middleware.Auth(cfg.JWTSecret), favoriteHandler.Add)
			properties.DELETE("/:id/favorites", middleware.Auth(cfg.JWTSecret), favoriteHandler.Remove)

			properties.POST("/:id/visits", middleware.Auth(cfg.JWTSecret), visitHandler.Request)

			properties.POST("",
				middleware.Auth(cfg.JWTSecret),
				middleware.RequireRole("AGENT", "DIRECTOR", "HQ", "SELLER"),
				propertyHandler.Create,
			)

			staff := properties.Group("")
			staff.Use(
				middleware.Auth(cfg.JWTSecret),
				middleware.RequireRole("AGENT", "DIRECTOR", "HQ"),
			)
			{
				staff.PUT("/:id", propertyHandler.Update)
				staff.DELETE("/:id", propertyHandler.Delete)

				staff.POST("/:id/validate", propertyHandler.Validate)

				staff.POST("/:id/photos", photoHandler.Add)
				staff.POST("/:id/photos/upload", photoHandler.Upload)
				staff.DELETE("/:id/photos/:photoId", photoHandler.Delete)
			}
		}

		api.GET("/favorites", middleware.Auth(cfg.JWTSecret), favoriteHandler.List)

		meGroup := api.Group("/me")
		meGroup.Use(middleware.Auth(cfg.JWTSecret))
		{
			meGroup.GET("/properties", middleware.RequireRole("AGENT", "DIRECTOR", "HQ"), propertyHandler.ListMine)
			meGroup.DELETE("", userHandler.DeleteMe)
		}

		mgmt := api.Group("/management")
		mgmt.Use(middleware.Auth(cfg.JWTSecret))
		{
			mgmt.GET("/properties", middleware.RequireRole("DIRECTOR", "HQ"), propertyHandler.ListAll)
			mgmt.GET("/pending-properties", middleware.RequireRole("AGENT", "DIRECTOR", "HQ"), propertyHandler.ListPending)
			mgmt.GET("/collaborators", middleware.RequireRole("DIRECTOR", "HQ", "IT"), userHandler.ListCollaborators)
			mgmt.GET("/users", middleware.RequireRole("IT", "HQ"), userHandler.ListAll)
			mgmt.DELETE("/users/:id", middleware.RequireRole("IT", "HQ"), userHandler.DeleteUser)
			mgmt.GET("/permissions", middleware.RequireRole("IT", "HQ"), permissionHandler.GetMatrix)

			staffReview := middleware.RequireRole("AGENT", "DIRECTOR", "HQ")
			mgmt.GET("/seller-applications", staffReview, sellerAppHandler.ListPending)
			mgmt.POST("/seller-applications/:id/approve", staffReview, sellerAppHandler.Approve)
			mgmt.POST("/seller-applications/:id/reject", staffReview, sellerAppHandler.Reject)
		}

		sellerApps := api.Group("/seller-applications")
		sellerApps.Use(middleware.Auth(cfg.JWTSecret))
		{
			sellerApps.POST("", sellerAppHandler.Apply)
			sellerApps.GET("", sellerAppHandler.Mine)
		}

		aiProxy := handlers.NewAIProxy(cfg.AIBaseURL)
		aiGroup := api.Group("/ai")
		{
			aiGroup.POST("/estimate", aiProxy)
			aiGroup.POST("/predict-delay", aiProxy)
			staffAI := aiGroup.Group("")
			staffAI.Use(middleware.Auth(cfg.JWTSecret), middleware.RequireRole("AGENT", "DIRECTOR", "HQ", "IT"))
			{
				staffAI.GET("/dashboard/kpis", aiProxy)
				staffAI.GET("/trends", aiProxy)
				staffAI.GET("/zones", aiProxy)
				staffAI.GET("/popular", aiProxy)
			}
		}

		messaging := api.Group("")
		messaging.Use(middleware.Auth(cfg.JWTSecret))
		{
			messaging.POST("/conversations", messageHandler.StartConversation)
			messaging.GET("/conversations", messageHandler.ListConversations)
			messaging.DELETE("/conversations/:id", messageHandler.Delete)
			messaging.POST("/conversations/:id/messages", messageHandler.SendMessage)
			messaging.GET("/conversations/:id/messages", messageHandler.GetMessages)
			messaging.GET("/messages/unread-count", messageHandler.UnreadCount)
		}

		visits := api.Group("")
		visits.Use(middleware.Auth(cfg.JWTSecret))
		{
			visits.GET("/visits", visitHandler.List)
			visits.PATCH("/visits/:id", visitHandler.UpdateStatus)
		}

		meetings := api.Group("/meetings")
		meetings.Use(middleware.Auth(cfg.JWTSecret))
		{
			meetings.GET("", meetingHandler.List)
			meetings.DELETE("/:id", meetingHandler.Delete)

			staffMeetings := meetings.Group("")
			staffMeetings.Use(middleware.RequireRole("AGENT", "DIRECTOR", "HQ"))
			{
				staffMeetings.POST("", meetingHandler.Create)
			}
		}

		alerts := api.Group("/alerts")
		alerts.Use(middleware.Auth(cfg.JWTSecret))
		{
			alerts.POST("", alertHandler.Create)
			alerts.GET("", alertHandler.List)
			alerts.DELETE("/:id", alertHandler.Delete)
		}

		sales := api.Group("/sales")
		sales.Use(middleware.Auth(cfg.JWTSecret))
		{
			sales.GET("", saleHandler.List)
			sales.GET("/:id", saleHandler.Get)

			staffSales := sales.Group("")
			staffSales.Use(middleware.RequireRole("AGENT", "DIRECTOR", "HQ"))
			{
				staffSales.POST("", saleHandler.Create)
				staffSales.PATCH("/:id", saleHandler.Update)
			}
		}
	}

	return r
}
