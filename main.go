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
	"highload-microservice/internal/redis"
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

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	// Initialize services
	userService := services.NewUserService(db, redisClient, kafkaProducer, logger)
	eventService := services.NewEventService(db, redisClient, kafkaProducer, logger)

	// Initialize worker pool for background processing
	workerPool := worker.NewPool(10, logger) // 10 workers
	workerPool.Start()

	// Add event processing job to worker pool
	workerPool.AddJob(func() {
		eventService.ProcessEvents(kafkaConsumer)
	})

	// Setup HTTP server
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	eventHandler := handlers.NewEventHandler(eventService, logger)

	// Setup routes
	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.GET("/", userHandler.ListUsers)
		}

		events := api.Group("/events")
		{
			events.POST("/", eventHandler.CreateEvent)
			events.GET("/", eventHandler.ListEvents)
			events.GET("/:id", eventHandler.GetEvent)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Start server in a goroutine
	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
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
