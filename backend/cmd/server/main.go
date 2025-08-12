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
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/saurabh22suman/oreo.io/internal/auth"
	"github.com/saurabh22suman/oreo.io/internal/database"
	"github.com/saurabh22suman/oreo.io/internal/handlers"
	"github.com/saurabh22suman/oreo.io/internal/middleware"
	"github.com/saurabh22suman/oreo.io/internal/repository"
	"github.com/saurabh22suman/oreo.io/internal/services"
)

func main() {
	// Load environment variables only if not in Docker
	// In Docker, environment variables are set by docker-compose
	if os.Getenv("DB_HOST") == "" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	} else {
		log.Println("Running in Docker - using environment variables from docker-compose")
	}

	// Initialize database connection - force real DB for projects functionality
	dbConn, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Initialize Redis connection with fallback to mock
	redisConn, err := database.NewRedisConnectionWithFallback()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		if closer, ok := redisConn.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	// Initialize services with real database
	log.Println("Using real database for all operations")

	// Create sqlx DB wrapper for project handlers
	sqlxDB := sqlx.NewDb(dbConn, "postgres")

	userRepo := repository.NewUserRepository(dbConn)
	projectHandlers := handlers.NewProjectHandlers(sqlxDB)
	log.Printf("Project handlers initialized: %+v", projectHandlers)
	if projectHandlers == nil {
		log.Fatal("Project handlers is nil!")
	}

	jwtService := auth.NewJWTService(os.Getenv("JWT_SECRET"))
	authService := services.NewAuthService(userRepo, jwtService)
	authHandlers := handlers.NewAuthHandlers(authService)
	sampleDataHandlers := handlers.NewSampleDataHandlers() // Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Set max multipart memory to 50MB (default is 32MB)
	router.MaxMultipartMemory = 50 << 20 // 50MB

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiting middleware
	router.Use(middleware.RateLimit())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"database":  "connected (mock in development)",
			"redis":     "connected (mock in development)",
		})
	})
	router.GET("/health/db", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"type":   "database",
		})
	})
	router.GET("/health/redis", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"type":   "redis",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Sample data routes (public)
		sampleData := v1.Group("/sample-data")
		{
			sampleData.GET("", sampleDataHandlers.ListSampleDatasets)
			sampleData.GET("/:category/:filename/info", sampleDataHandlers.GetSampleDatasetInfo)
			sampleData.GET("/:category/:filename/download", sampleDataHandlers.DownloadSampleDataset)
			sampleData.GET("/:category/:filename/preview", sampleDataHandlers.PreviewSampleDataset)
		}

		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandlers.RegisterWithService())
			auth.POST("/login", authHandlers.LoginWithService())
			auth.POST("/refresh", authHandlers.RefreshTokenWithService())
			auth.POST("/logout", handlers.Logout())
			auth.GET("/me", middleware.RequireAuthWithService(authService), handlers.GetCurrentUser())
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.RequireAuthWithService(authService))
		{
			// Project routes
			log.Printf("Registering project routes with handlers: %+v", projectHandlers)
			projects := protected.Group("/projects")
			{
				projects.GET("", projectHandlers.GetProjects())
				projects.POST("", projectHandlers.CreateProject())
				projects.GET("/:id", projectHandlers.GetProject())
				projects.PUT("/:id", projectHandlers.UpdateProject())
				projects.DELETE("/:id", projectHandlers.DeleteProject())
			}

			// Dataset routes
			datasetHandlers := handlers.NewDatasetHandlers(sqlxDB)
			datasets := protected.Group("/datasets")
			{
				datasets.POST("/upload", datasetHandlers.UploadDataset())
				datasets.GET("/user", datasetHandlers.GetUserDatasets())
				datasets.GET("/project/:project_id", datasetHandlers.GetDatasets())
				datasets.GET("/:id", datasetHandlers.GetDatasetByID())
				datasets.DELETE("/:id", datasetHandlers.DeleteDataset())
			}

			// Schema routes
			schemaRepo := repository.NewSchemaRepository(sqlxDB)
			schemaHandlers := handlers.NewSchemaHandlers(sqlxDB)
			schemas := protected.Group("/schemas")
			{
				schemas.POST("", schemaHandlers.CreateSchema())
				schemas.GET("/dataset/:dataset_id", schemaHandlers.GetSchema())
				schemas.POST("/infer/:dataset_id", schemaHandlers.InferSchema()) // Schema inference endpoint
				schemas.PUT("/:schema_id", schemaHandlers.UpdateSchema())
				schemas.DELETE("/:schema_id", schemaHandlers.DeleteSchema())
			}

			// Data routes
			data := protected.Group("/data")
			{
				data.GET("/dataset/:dataset_id", schemaHandlers.GetDatasetData())
				data.POST("/dataset/:dataset_id/query", schemaHandlers.QueryDatasetData())
				data.PUT("/dataset/:dataset_id", schemaHandlers.UpdateDatasetData())
				data.DELETE("/dataset/:dataset_id/row/:row_index", schemaHandlers.DeleteDatasetData())
			}

			// Data submission routes for append functionality
			submissionRepo := repository.NewDataSubmissionRepository(sqlxDB)
			validationSvc := services.NewValidationService(schemaRepo, submissionRepo)
			submissionHandlers := handlers.NewDataSubmissionHandlers(submissionRepo, schemaRepo, validationSvc)
			
			// User submission routes
			datasets.POST("/:dataset_id/append", submissionHandlers.SubmitDataForAppend())
			datasets.GET("/:dataset_id/submissions", submissionHandlers.GetDataSubmissions())
			
			// Submission management routes
			submissions := protected.Group("/submissions")
			{
				submissions.GET("/:submission_id/details", submissionHandlers.GetSubmissionDetails())
			}
			
			// Staging data routes for live editing
			staging := protected.Group("/staging")
			{
				staging.PUT("/:staging_id", submissionHandlers.UpdateStagingData())
			}

			// Business rules routes
			businessRules := protected.Group("/datasets/:dataset_id/rules")
			{
				businessRules.POST("", submissionHandlers.CreateBusinessRule())
				businessRules.GET("", submissionHandlers.GetBusinessRules())
			}

			// Admin routes for submission review
			admin := protected.Group("/admin")
			{
				admin.GET("/submissions/pending", submissionHandlers.GetPendingSubmissions())
				admin.PUT("/submissions/:submission_id/review", submissionHandlers.ReviewSubmission())
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
