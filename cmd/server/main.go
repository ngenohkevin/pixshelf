package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	// Set up the API endpoints
	imageHandler := handlers.NewImageHandler(imageService)
	imageHandler.RegisterRoutes(router)

	// Set up the UI endpoints
	uiHandler := ui.NewUIHandler(imageService)
	uiHandler.RegisterRoutes(router)

	// Serve static files
	router.Static("/static", "./static")

	// Set up the server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s in %s mode", addr, cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
