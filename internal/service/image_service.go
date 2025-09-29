package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngenohkevin/pixshelf/internal/config"
	"github.com/ngenohkevin/pixshelf/internal/models"
	"github.com/ngenohkevin/pixshelf/internal/repository"
)

// ImageService handles business logic for images
type ImageService struct {
	repo        *repository.ImageRepository
	cfg         *config.Config
	uploadPath  string
	maxFileSize int64
}

// GetUploadPath returns the path where images are stored
func (s *ImageService) GetUploadPath() string {
	return s.uploadPath
}

// NewImageService creates a new ImageService
func NewImageService(repo *repository.ImageRepository, cfg *config.Config) *ImageService {
	return &ImageService{
		repo:        repo,
		cfg:         cfg,
		uploadPath:  cfg.ImageStorage,
		maxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// GetByID retrieves an image by ID for a specific user
func (s *ImageService) GetByID(ctx context.Context, id int64, userID int64) (*models.PublicImage, error) {
	img, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// List retrieves a paginated list of images for a specific user
func (s *ImageService) List(ctx context.Context, userID int64, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := &models.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	total, err := s.repo.Count(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	imgs, err := s.repo.List(ctx, userID, pagination)
	if err != nil {
		return nil, nil, err
	}

	publicImgs := make([]*models.PublicImage, len(imgs))
	for i, img := range imgs {
		publicImgs[i] = models.NewPublicImage(img, s.cfg.BaseURL)
	}

	return publicImgs, pagination, nil
}

// Search searches for images by name or description for a specific user
func (s *ImageService) Search(ctx context.Context, userID int64, query string, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := &models.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	total, err := s.repo.SearchCount(ctx, userID, query)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	params := &models.SearchParams{
		Query:      query,
		Pagination: pagination,
	}

	imgs, err := s.repo.Search(ctx, userID, params)
	if err != nil {
		return nil, nil, err
	}

	publicImgs := make([]*models.PublicImage, len(imgs))
	for i, img := range imgs {
		publicImgs[i] = models.NewPublicImage(img, s.cfg.BaseURL)
	}

	return publicImgs, pagination, nil
}

// Create creates a new image for a specific user
func (s *ImageService) Create(ctx context.Context, userID int64, fileHeader interface{}, name, description string) (*models.PublicImage, error) {
	file, ok := fileHeader.(*multipart.FileHeader)
	if !ok {
		return nil, errors.New("invalid file type")
	}

	log.Printf("Received file: %s, size: %d bytes", file.Filename, file.Size)

	if file.Size > s.maxFileSize {
		return nil, errors.New("file too large")
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Use the provided name or fall back to the original filename
	displayName := name
	if displayName == "" {
		// Remove extension from filename for display name
		displayName = strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	}

	// Generate a unique filename for storage using the provided name
	// Format: {timestamp}_{provided_name_sanitized}.{original_extension}
	originalExt := filepath.Ext(file.Filename)
	nameToUse := displayName
	if nameToUse == "" {
		nameToUse = strings.TrimSuffix(file.Filename, originalExt)
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(nameToUse), originalExt)
	filePath := filepath.Join(s.uploadPath, filename)

	log.Printf("Saving file to: %s", filePath)

	// Open the file for reading
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			return
		}
	}(src)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			return
		}
	}(dst)

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	log.Printf("File saved successfully to %s", filePath)

	// Create the image record
	img := &models.Image{
		Name:        displayName,
		Description: description,
		FilePath:    filename,
		MimeType:    file.Header.Get("Content-Type"),
		SizeBytes:   file.Size,
		UserID:      &userID,
	}

	// Save to database
	img, err = s.repo.Create(ctx, img)
	if err != nil {
		// Clean up the file if database insertion fails
		err := os.Remove(filePath)
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// Update updates an image's metadata for a specific user
func (s *ImageService) Update(ctx context.Context, id int64, userID int64, name, description string) (*models.PublicImage, error) {
	// Check if image exists and belongs to user
	img, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update image metadata
	img.Name = name
	img.Description = description

	// Save to database
	img, err = s.repo.Update(ctx, img, userID)
	if err != nil {
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// Delete deletes an image for a specific user
func (s *ImageService) Delete(ctx context.Context, id int64, userID int64) error {
	// Get the image to retrieve its file path and verify ownership
	img, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Delete the image file
	filePath := filepath.Join(s.uploadPath, img.FilePath)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete image file: %w", err)
	}

	// Delete from database
	return s.repo.Delete(ctx, id, userID)
}

// Helper functions
func sanitizeFilename(filename string) string {
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")
	// Remove any path separators
	filename = filepath.Base(filename)
	// Only keep alphanumeric characters, underscores, dashes, and dots
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, filename)
}
