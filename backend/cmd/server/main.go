package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/saurabh22suman/oreo.io/internal/database"
	"github.com/saurabh22suman/oreo.io/internal/handlers"
	"github.com/saurabh22suman/oreo.io/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redis, err := database.NewRedisConnection()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiting middleware
	router.Use(middleware.RateLimit())

	// Health check endpoints
	router.GET("/health", handlers.HealthCheck(db, redis))
	router.GET("/health/db", handlers.DatabaseHealthCheck(db))
	router.GET("/health/redis", handlers.RedisHealthCheck(redis))

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes will be added here
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register())
			auth.POST("/login", handlers.Login())
			auth.POST("/logout", handlers.Logout())
			auth.GET("/me", middleware.RequireAuth(), handlers.GetCurrentUser())
		}

		// Protected routes will be added here
		protected := v1.Group("")
		protected.Use(middleware.RequireAuth())
		{
			// Project routes
			projects := protected.Group("/projects")
			{
				projects.GET("", handlers.GetProjects())
				projects.POST("", handlers.CreateProject())
				projects.GET("/:id", handlers.GetProject())
				projects.PUT("/:id", handlers.UpdateProject())
				projects.DELETE("/:id", handlers.DeleteProject())
			}
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a 5-second timeout to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
