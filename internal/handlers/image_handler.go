package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngenohkevin/pixshelf/internal/auth"
	"github.com/ngenohkevin/pixshelf/internal/db/sqlc"
	"github.com/ngenohkevin/pixshelf/internal/models"
	"github.com/ngenohkevin/pixshelf/internal/service"
	"github.com/ngenohkevin/pixshelf/internal/utils"
	"github.com/ngenohkevin/pixshelf/templates"
)

// ImageHandler handles HTTP requests for images
type ImageHandler struct {
	service   *service.ImageService
	db        *sqlc.Queries
	optimizer *service.ImageOptimizer
}

// NewImageHandler creates a new ImageHandler
func NewImageHandler(service *service.ImageService, db *sqlc.Queries, optimizer *service.ImageOptimizer) *ImageHandler {
	return &ImageHandler{
		service:   service,
		db:        db,
		optimizer: optimizer,
	}
}

// GetImage retrieves an image by ID
func (h *ImageHandler) GetImage(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, fmt.Errorf("invalid image ID: %w", err))
		return
	}

	img, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	c.JSON(http.StatusOK, img)
}

// ListImages retrieves a list of images
func (h *ImageHandler) ListImages(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	imgs, pagination, err := h.service.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images":     imgs,
		"pagination": pagination,
	})
}

// SearchImages searches for images by name or description
func (h *ImageHandler) SearchImages(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, fmt.Errorf("search query is required"))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	imgs, pagination, err := h.service.Search(c.Request.Context(), userID, query, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images":     imgs,
		"pagination": pagination,
	})
}

// UploadImage uploads a new image
func (h *ImageHandler) UploadImage(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	log.Println("Upload image handler called")

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")

	log.Printf("Name: %s, Description: %s", name, description)

	// Get the file
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		utils.BadRequest(c, fmt.Errorf("image is required: %w", err))
		return
	}

	log.Printf("File received: %s, size: %d", file.Filename, file.Size)

	// Create the image
	_, err = h.service.Create(c.Request.Context(), userID, file, name, description)
	if err != nil {
		log.Printf("Error creating image: %v", err)
		utils.InternalServerError(c, err)
		return
	}

	// Redirect to home page on success
	c.Redirect(http.StatusSeeOther, "/")
}

// UpdateImage updates an image's metadata
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, fmt.Errorf("invalid image ID: %w", err))
		return
	}

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")

	// Validate input
	if name == "" {
		utils.BadRequest(c, fmt.Errorf("name is required"))
		return
	}

	// Update the image
	img, err := h.service.Update(c.Request.Context(), id, userID, name, description)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Get current user data for template
		sqlcUser, err := auth.GetCurrentUser(c, h.db)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}
		user := auth.ConvertUserToTemplateData(sqlcUser)

		// Convert image to template data
		imageData := &templates.ImageData{
			ID:          img.ID,
			Name:        img.Name,
			Description: img.Description,
			URL:         img.URL,
			PublicURL:   img.PublicURL,
			MimeType:    img.MimeType,
			SizeBytes:   img.SizeBytes,
			CreatedAt:   img.CreatedAt,
		}

		// Render the image detail template
		component := templates.ImageDetail(imageData, user)
		component.Render(c.Request.Context(), c.Writer)
		return
	}

	// For non-HTMX requests, return JSON
	c.JSON(http.StatusOK, img)
}

// DeleteImage deletes an image
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, fmt.Errorf("invalid image ID: %w", err))
		return
	}

	err = h.service.Delete(c.Request.Context(), id, userID)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	// Add HX-Redirect header to ensure redirection to gallery
	c.Header("HX-Redirect", "/")
	c.Status(http.StatusNoContent)
}

// GetImageByFilePath retrieves an image by its file path
func (h *ImageHandler) GetImageByFilePath(c *gin.Context) {
	filePath := c.Param("filepath")
	if filePath == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(h.service.GetUploadPath(), filePath)

	// Cloudflare-optimized headers
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Vary", "Accept-Encoding")

	// Generate simple ETag from file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
	c.Header("ETag", etag)

	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	// Serve the static file
	c.File(fullPath)
}

// GetImageVariant serves an image variant (resized version)
func (h *ImageHandler) GetImageVariant(c *gin.Context) {
	size := c.Param("size")
	filePath := c.Param("filepath")

	// Remove leading slash from filepath since it's captured as /*filepath
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	if filePath == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	originalPath := filepath.Join(h.service.GetUploadPath(), filePath)

	// Check if original exists
	if _, err := os.Stat(originalPath); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	widthMap := map[string]int{
		"thumb":  150,
		"small":  480,
		"medium": 800,
	}

	// Serve original if requested
	if size == "original" {
		// Cloudflare-optimized headers
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("Vary", "Accept-Encoding")

		// Generate ETag
		fileInfo, _ := os.Stat(originalPath)
		etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
		c.Header("ETag", etag)

		if match := c.GetHeader("If-None-Match"); match == etag {
			c.Status(http.StatusNotModified)
			return
		}

		c.File(originalPath)
		return
	}

	// Get width for the requested size
	width, ok := widthMap[size]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Get or create variant
	variantPath, err := h.optimizer.GetOrCreateVariant(originalPath, width)
	if err != nil {
		// Fallback to original on error
		log.Printf("Error creating variant: %v", err)
		c.File(originalPath)
		return
	}

	// Cloudflare-optimized headers
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Vary", "Accept-Encoding")

	// Generate ETag for variant
	fileInfo, _ := os.Stat(variantPath)
	etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
	c.Header("ETag", etag)

	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.File(variantPath)
}

// RegisterRoutes registers the image routes
func (h *ImageHandler) RegisterRoutes(router gin.IRouter) {
	api := router.Group("/api")
	{
		api.GET("/images", h.ListImages)
		api.GET("/images/search", h.SearchImages)
		api.GET("/images/:id", h.GetImage)
		api.POST("/images", h.UploadImage)
		api.PUT("/images/:id", h.UpdateImage)
		api.DELETE("/images/:id", h.DeleteImage)
	}

	// Note: public-images route is now handled in main.go as a public route
}

// ImageService defines the interface for image service
type ImageService interface {
	GetByID(ctx context.Context, id int64, userID int64) (*models.PublicImage, error)
	List(ctx context.Context, userID int64, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error)
	Search(ctx context.Context, userID int64, query string, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error)
	Create(ctx context.Context, userID int64, file interface{}, name, description string) (*models.PublicImage, error)
	Update(ctx context.Context, id int64, userID int64, name, description string) (*models.PublicImage, error)
	Delete(ctx context.Context, id int64, userID int64) error
	GetUploadPath() string
}
