package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ngenohkevin/pixshelf/internal/db/sqlc"
	"github.com/ngenohkevin/pixshelf/internal/models"
)

// ImageRepository handles database operations for images
type ImageRepository struct {
	q sqlc.Querier
}

// NewImageRepository creates a new ImageRepository
func NewImageRepository(q sqlc.Querier) *ImageRepository {
	return &ImageRepository{q: q}
}

// GetByID retrieves an image by ID
func (r *ImageRepository) GetByID(ctx context.Context, id int64) (*models.Image, error) {
	img, err := r.q.GetImage(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// List retrieves a paginated list of images
func (r *ImageRepository) List(ctx context.Context, pagination *models.Pagination) ([]*models.Image, error) {
	arg := sqlc.ListImagesParams{
		Limit:  int32(pagination.PageSize),
		Offset: int32((pagination.Page - 1) * pagination.PageSize),
	}

	imgs, err := r.q.ListImages(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return convertSQLCImages(imgs), nil
}

// Count returns the total number of images
func (r *ImageRepository) Count(ctx context.Context) (int, error) {
	count, err := r.q.CountImages(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return int(count), nil
}

// Search searches for images by name or description
func (r *ImageRepository) Search(ctx context.Context, params *models.SearchParams) ([]*models.Image, error) {
	pattern := "%" + params.Query + "%"
	arg := sqlc.SearchImagesParams{
		Name:   pattern,
		Limit:  int32(params.Pagination.PageSize),
		Offset: int32((params.Pagination.Page - 1) * params.Pagination.PageSize),
	}

	imgs, err := r.q.SearchImages(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to search images: %w", err)
	}

	return convertSQLCImages(imgs), nil
}

// SearchCount returns the total number of images matching a search query
func (r *ImageRepository) SearchCount(ctx context.Context, query string) (int, error) {
	pattern := "%" + query + "%"
	count, err := r.q.CountSearchImages(ctx, pattern)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return int(count), nil
}

// Create creates a new image
func (r *ImageRepository) Create(ctx context.Context, image *models.Image) (*models.Image, error) {
	var description pgtype.Text
	description.String = image.Description
	description.Valid = image.Description != ""

	arg := sqlc.CreateImageParams{
		Name:        image.Name,
		Description: description,
		FilePath:    image.FilePath,
		MimeType:    image.MimeType,
		SizeBytes:   image.SizeBytes,
	}

	img, err := r.q.CreateImage(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// Update updates an image's metadata
func (r *ImageRepository) Update(ctx context.Context, image *models.Image) (*models.Image, error) {
	var description pgtype.Text
	description.String = image.Description
	description.Valid = image.Description != ""

	arg := sqlc.UpdateImageParams{
		ID:          int32(image.ID),
		Name:        image.Name,
		Description: description,
	}

	img, err := r.q.UpdateImage(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// Delete deletes an image
func (r *ImageRepository) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteImage(ctx, int32(id)); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// Helper functions to convert between SQLC and domain models
func convertSQLCImage(img sqlc.Image) *models.Image {
	description := ""
	if img.Description.Valid {
		description = img.Description.String
	}

	createdAt := time.Now()
	if img.CreatedAt.Valid {
		createdAt = img.CreatedAt.Time
	}

	updatedAt := time.Now()
	if img.UpdatedAt.Valid {
		updatedAt = img.UpdatedAt.Time
	}

	return &models.Image{
		ID:          int64(img.ID),
		Name:        img.Name,
		Description: description,
		FilePath:    img.FilePath,
		MimeType:    img.MimeType,
		SizeBytes:   img.SizeBytes,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func convertSQLCImages(imgs []sqlc.Image) []*models.Image {
	result := make([]*models.Image, len(imgs))
	for i, img := range imgs {
		result[i] = convertSQLCImage(img)
	}
	return result
}
