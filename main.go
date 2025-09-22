package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"highload-microservice/internal/config"
	"highload-microservice/internal/database"
	"highload-microservice/internal/handlers"
	"highload-microservice/internal/kafka"
	"highload-microservice/internal/middleware"
	"highload-microservice/internal/models"
	"highload-microservice/internal/redis"
	"highload-microservice/internal/security"
	"highload-microservice/internal/services"
	"highload-microservice/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.LogLevel == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Validate secrets
	if errors := config.ValidateSecrets(cfg); len(errors) > 0 {
		logger.Warn("Security issues found in configuration:")
		for _, err := range errors {
			logger.Warnf("  - %s", err)
		}
		logger.Warn("Use 'go run cmd/secrets/main.go validate' to check secrets")
		logger.Warn("Use 'go run cmd/secrets/main.go set <key>' to set secure values")
	}

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("Database migrations completed successfully")

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize Kafka
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		logger.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	kafkaConsumer, err := kafka.NewConsumer(cfg.Kafka)
	if err != nil {
		logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// Initialize security auditor
	securityAuditor := security.NewSecurityAuditor(logger)

	// Initialize services
	userService := services.NewUserService(db, redisClient, kafkaProducer, logger)
	eventService := services.NewEventService(db, redisClient, kafkaProducer, logger)

	// Initialize auth service
	authConfig := services.AuthConfig{
		JWTSecret:         cfg.Auth.JWTSecret,
		JWTExpiration:     time.Duration(cfg.Auth.JWTExpiration) * time.Hour,
		RefreshExpiration: time.Duration(cfg.Auth.RefreshExpiration) * 24 * time.Hour,
		APIKeyLength:      cfg.Auth.APIKeyLength,
	}
	authService := services.NewAuthService(db, logger, authConfig)

	// Initialize worker pool for background processing
	workerPool := worker.NewPool(10, logger) // 10 workers
	workerPool.Start()

	// Add event processing job to worker pool
	workerPool.AddJob(func() {
		eventService.ProcessEvents(kafkaConsumer)
	})

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	eventHandler := handlers.NewEventHandler(eventService, logger)
	authHandler := handlers.NewAuthHandler(authService, securityAuditor, logger)
	securityHandler := handlers.NewSecurityHandler(securityAuditor, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	validationMiddleware := middleware.NewValidationMiddleware(logger)
	securityLoggingMiddleware := middleware.NewSecurityLoggingMiddleware(securityAuditor, logger)

	// Initialize security middleware
	securityConfig := middleware.SecurityConfig{
		AllowedOrigins:        cfg.Security.AllowedOrigins,
		AllowedMethods:        cfg.Security.AllowedMethods,
		AllowedHeaders:        cfg.Security.AllowedHeaders,
		ExposedHeaders:        cfg.Security.ExposedHeaders,
		AllowCredentials:      cfg.Security.AllowCredentials,
		MaxAge:                cfg.Security.MaxAge,
		ContentTypeNosniff:    cfg.Security.ContentTypeNosniff,
		FrameDeny:             cfg.Security.FrameDeny,
		XSSProtection:         cfg.Security.XSSProtection,
		ReferrerPolicy:        cfg.Security.ReferrerPolicy,
		PermissionsPolicy:     cfg.Security.PermissionsPolicy,
		ContentSecurityPolicy: cfg.Security.ContentSecurityPolicy,
	}
	securityMiddleware := middleware.NewSecurityMiddleware(securityConfig, logger)

	// Setup HTTP server
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Apply security middleware globally
	router.Use(securityMiddleware.RequestID())
	router.Use(securityMiddleware.SecurityHeaders())
	router.Use(securityMiddleware.SecurityLogging())
	router.Use(securityMiddleware.CORS())

	// Apply security logging middleware
	router.Use(securityLoggingMiddleware.LogRequest())
	router.Use(securityLoggingMiddleware.LogSuspiciousInput())

	// Initialize rate limiting middleware
	var rateLimitMiddleware *middleware.RateLimitMiddleware
	if cfg.RateLimit.Enabled {
		rateLimitConfig := middleware.RateLimitConfig{
			Requests: cfg.RateLimit.RequestsPerMinute,
			Duration: 1 * time.Minute,
		}
		rateLimitMiddleware = middleware.NewRateLimitMiddleware(rateLimitConfig, logger)
	}

    // Initialize DDoS protection (can be disabled via env for CI)
    ddosConfig := middleware.DDoSConfig{
        MaxRequests:     100,
        WindowDuration:  1 * time.Minute,
        BlockDuration:   5 * time.Minute,
        CleanupInterval: 1 * time.Minute,
    }
    ddosProtection := middleware.NewDDoSProtection(ddosConfig, logger)
    ddosEnabled := os.Getenv("DDOS_PROTECTION_ENABLED")

	// Setup routes
	api := router.Group("/api/v1")
	{
        // Apply DDoS protection to all API routes unless disabled
        if ddosEnabled != "false" {
            api.Use(ddosProtection.Protect())
        }

		// Apply input sanitization to all API routes
		api.Use(validationMiddleware.SanitizeInput())

		// Apply rate limiting to all API routes if enabled
		if rateLimitMiddleware != nil {
			api.Use(rateLimitMiddleware.RateLimit())
		}

		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			// Apply strict rate limiting to auth endpoints
			if rateLimitMiddleware != nil {
				auth.Use(rateLimitMiddleware.AuthRateLimit())
			}

			auth.POST("/login", validationMiddleware.ValidateRequest(&models.LoginRequest{}), authHandler.Login)
			auth.POST("/refresh", validationMiddleware.ValidateRequest(&models.RefreshTokenRequest{}), authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
			auth.GET("/profile", authMiddleware.RequireAuth(), authHandler.GetProfile)
		}

		// API Key management (admin only)
		apiKeys := api.Group("/api-keys")
		apiKeys.Use(authMiddleware.RequireAuth(), authMiddleware.RequireRole("admin"))
		{
			apiKeys.POST("/", validationMiddleware.ValidateRequest(&models.CreateAPIKeyRequest{}), authHandler.CreateAPIKey)
		}

		// User management routes (authenticated)
		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.POST("/", authMiddleware.RequireRole("admin"), validationMiddleware.ValidateRequest(&models.CreateUserRequest{}), userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", validationMiddleware.ValidateRequest(&models.UpdateUserRequest{}), userHandler.UpdateUser)
			users.DELETE("/:id", authMiddleware.RequireRole("admin"), userHandler.DeleteUser)
			users.GET("/", validationMiddleware.ValidatePagination(), userHandler.ListUsers)
		}

		// Event management routes (authenticated)
		events := api.Group("/events")
		events.Use(authMiddleware.RequireAuth())
		{
			events.POST("/", validationMiddleware.ValidateRequest(&models.CreateEventRequest{}), eventHandler.CreateEvent)
			events.GET("/", validationMiddleware.ValidatePagination(), eventHandler.ListEvents)
			events.GET("/:id", eventHandler.GetEvent)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(503, gin.H{
				"status":    "unhealthy",
				"error":     "database connection failed",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Check Redis connection
		if err := redisClient.Ping(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{
				"status":    "unhealthy",
				"error":     "redis connection failed",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// DDoS protection stats endpoint (admin only)
	router.GET("/admin/ddos-stats", authMiddleware.RequireAuth(), authMiddleware.RequireRole("admin"), func(c *gin.Context) {
		stats := ddosProtection.GetStats()
		c.JSON(200, gin.H{
			"ddos_protection": stats,
			"timestamp":       time.Now().Unix(),
		})
	})

	// Security monitoring endpoints (admin only)
	securityAdmin := router.Group("/admin/security")
	securityAdmin.Use(authMiddleware.RequireAuth(), authMiddleware.RequireRole("admin"))
	{
		securityAdmin.GET("/stats", securityHandler.GetSecurityStats)
		securityAdmin.GET("/alerts", securityHandler.GetSecurityAlerts)
		securityAdmin.GET("/events", securityHandler.GetSecurityEvents)
		securityAdmin.GET("/threats", securityHandler.GetThreatIntelligence)
		securityAdmin.GET("/health", securityHandler.GetSecurityHealth)
	}

	// Start server in a goroutine
	server := &http.Server{
		Addr:              cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, // Prevent Slowloris attacks
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if cfg.Server.UseTLS {
			logger.Infof("Starting HTTPS server on %s:%s", cfg.Server.Host, cfg.Server.Port)
			if err := server.ListenAndServeTLS(cfg.Server.TLSCert, cfg.Server.TLSKey); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			logger.Infof("Starting HTTP server on %s:%s", cfg.Server.Host, cfg.Server.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Stop worker pool
	workerPool.Stop()

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
