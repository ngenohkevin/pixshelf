package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/ngenohkevin/pixshelf/internal/auth"
	"github.com/ngenohkevin/pixshelf/internal/config"
	"github.com/ngenohkevin/pixshelf/internal/db"
	"github.com/ngenohkevin/pixshelf/internal/db/sqlc"
	"github.com/ngenohkevin/pixshelf/internal/handlers"
	"github.com/ngenohkevin/pixshelf/internal/handlers/ui"
	"github.com/ngenohkevin/pixshelf/internal/repository"
	"github.com/ngenohkevin/pixshelf/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set the Gin mode based on the environment
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Set up the database connection
	dbPool, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Initialize the SQLC queries
	queries := sqlc.New(dbPool)

	// Initialize the repository
	imageRepo := repository.NewImageRepository(queries)

	// Initialize the service
	imageService := service.NewImageService(imageRepo, cfg)

	// Set up the Gin router
	router := gin.Default()

	// Set up sessions
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   !cfg.IsDevelopment(),
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("pixshelf_session", store))

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Set up CORS if needed
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Initialize auth service
	authConfig := &auth.AuthConfig{
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		BaseURL:            cfg.BaseURL,
	}
	authService := auth.NewAuthService(authConfig, queries)
	authHandler := auth.NewAuthHandler(authService)

	// Set up auth routes (these don't require authentication)
	authHandler.RegisterRoutes(router)

	// Initialize handlers for both protected and public routes
	imageHandler := handlers.NewImageHandler(imageService, queries)

	// Public routes (no authentication required)
	public := router.Group("/")
	{
		// Public image access for external apps
		public.GET("/public-images/:filepath", imageHandler.GetImageByFilePath)
	}

	// Protected routes
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		// Set up the API endpoints
		imageHandler.RegisterRoutes(protected)

		// Set up the UI endpoints
		uiHandler := ui.NewUIHandler(imageService, queries)
		uiHandler.RegisterRoutes(protected)

		// Serve static files
		protected.Static("/static", "./static")
	}

	// Set up the server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s in %s mode", addr, cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server up to 5 seconds to shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
