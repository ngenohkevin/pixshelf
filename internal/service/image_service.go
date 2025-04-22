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

// NewImageService creates a new ImageService
func NewImageService(repo *repository.ImageRepository, cfg *config.Config) *ImageService {
	return &ImageService{
		repo:        repo,
		cfg:         cfg,
		uploadPath:  cfg.ImageStorage,
		maxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// GetByID retrieves an image by ID
func (s *ImageService) GetByID(ctx context.Context, id int64) (*models.PublicImage, error) {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// List retrieves a paginated list of images
func (s *ImageService) List(ctx context.Context, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error) {
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

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	imgs, err := s.repo.List(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	publicImgs := make([]*models.PublicImage, len(imgs))
	for i, img := range imgs {
		publicImgs[i] = models.NewPublicImage(img, s.cfg.BaseURL)
	}

	return publicImgs, pagination, nil
}

// Search searches for images by name or description
func (s *ImageService) Search(ctx context.Context, query string, page, pageSize int) ([]*models.PublicImage, *models.Pagination, error) {
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

	total, err := s.repo.SearchCount(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	params := &models.SearchParams{
		Query:      query,
		Pagination: pagination,
	}

	imgs, err := s.repo.Search(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	publicImgs := make([]*models.PublicImage, len(imgs))
	for i, img := range imgs {
		publicImgs[i] = models.NewPublicImage(img, s.cfg.BaseURL)
	}

	return publicImgs, pagination, nil
}

// Create creates a new image
func (s *ImageService) Create(ctx context.Context, fileHeader interface{}, name, description string) (*models.PublicImage, error) {
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

	// Generate a unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(file.Filename))
	filePath := filepath.Join(s.uploadPath, filename)

	log.Printf("Saving file to: %s", filePath)

	// Open the file for reading
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	log.Printf("File saved successfully to %s", filePath)

	// Create the image record
	img := &models.Image{
		Name:        name,
		Description: description,
		FilePath:    filename,
		MimeType:    file.Header.Get("Content-Type"),
		SizeBytes:   file.Size,
	}

	// Save to database
	img, err = s.repo.Create(ctx, img)
	if err != nil {
		// Clean up the file if database insertion fails
		os.Remove(filePath)
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// Update updates an image's metadata
func (s *ImageService) Update(ctx context.Context, id int64, name, description string) (*models.PublicImage, error) {
	// Check if image exists
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update image metadata
	img.Name = name
	img.Description = description

	// Save to database
	img, err = s.repo.Update(ctx, img)
	if err != nil {
		return nil, err
	}

	return models.NewPublicImage(img, s.cfg.BaseURL), nil
}

// Delete deletes an image
func (s *ImageService) Delete(ctx context.Context, id int64) error {
	// Get the image to retrieve its file path
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the image file
	filePath := filepath.Join(s.uploadPath, img.FilePath)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete image file: %w", err)
	}

	// Delete from database
	return s.repo.Delete(ctx, id)
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
