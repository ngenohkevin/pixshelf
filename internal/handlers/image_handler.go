package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngenohkevin/pixshelf/internal/models"
	"github.com/ngenohkevin/pixshelf/internal/service"
	"github.com/ngenohkevin/pixshelf/internal/utils"
)

// ImageHandler handles HTTP requests for images
type ImageHandler struct {
	service *service.ImageService
}

// NewImageHandler creates a new ImageHandler
func NewImageHandler(service *service.ImageService) *ImageHandler {
	return &ImageHandler{service: service}
}

// GetImage retrieves an image by ID
func (h *ImageHandler) GetImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, fmt.Errorf("invalid image ID: %w", err))
		return
	}

	img, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	c.JSON(http.StatusOK, img)
}

// ListImages retrieves a list of images
func (h *ImageHandler) ListImages(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	imgs, pagination, err := h.service.List(c.Request.Context(), page, pageSize)
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

	imgs, pagination, err := h.service.Search(c.Request.Context(), query, page, pageSize)
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
	log.Println("Upload image handler called")

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")

	log.Printf("Name: %s, Description: %s", name, description)

	// Validate name
	if name == "" {
		utils.BadRequest(c, fmt.Errorf("name is required"))
		return
	}

	// Get the file
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		utils.BadRequest(c, fmt.Errorf("image is required: %w", err))
		return
	}

	log.Printf("File received: %s, size: %d", file.Filename, file.Size)

	// Create the image
	_, err = h.service.Create(c.Request.Context(), file, name, description)
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
	img, err := h.service.Update(c.Request.Context(), id, name, description)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	c.JSON(http.StatusOK, img)
}

// DeleteImage deletes an image
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, fmt.Errorf("invalid image ID: %w", err))
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "Image", id)
		return
	}

	c.Status(http.StatusNoContent)
}

// RegisterRoutes registers the image routes
func (h *ImageHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/images", h.ListImages)
		api.GET("/images/search", h.SearchImages)
		api.GET("/images/:id", h.GetImage)
		api.POST("/images", h.UploadImage)
		api.PUT("/images/:id", h.UpdateImage)
		api.DELETE("/images/:id", h.DeleteImage)
	}
}

// ImageService defines the interface for image service
type ImageService interface {
	GetByID(ctx context.Context, id int64) (*models.PublicImage, error)
	List(ctx context.Context, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error)
	Create(ctx context.Context, file interface{}, name, description string) (*models.PublicImage, error)
	Update(ctx context.Context, id int64, name, description string) (*models.PublicImage, error)
	Delete(ctx context.Context, id int64) error
}
