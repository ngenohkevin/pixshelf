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

// GetByID retrieves an image by ID for a specific user
func (r *ImageRepository) GetByID(ctx context.Context, id int64, userID int64) (*models.Image, error) {
	img, err := r.q.GetImageByUser(ctx, sqlc.GetImageByUserParams{
		ID:     int32(id),
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// List retrieves a paginated list of images for a specific user
func (r *ImageRepository) List(ctx context.Context, userID int64, pagination *models.Pagination) ([]*models.Image, error) {
	arg := sqlc.ListImagesParams{
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
		Limit:  int32(pagination.PageSize),
		Offset: int32((pagination.Page - 1) * pagination.PageSize),
	}

	imgs, err := r.q.ListImages(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return convertSQLCImages(imgs), nil
}

// Count returns the total number of images for a specific user
func (r *ImageRepository) Count(ctx context.Context, userID int64) (int, error) {
	count, err := r.q.CountImages(ctx, pgtype.Int4{Int32: int32(userID), Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return int(count), nil
}

// Search searches for images by name or description for a specific user
func (r *ImageRepository) Search(ctx context.Context, userID int64, params *models.SearchParams) ([]*models.Image, error) {
	pattern := "%" + params.Query + "%"
	arg := sqlc.SearchImagesParams{
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
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

// SearchCount returns the total number of images matching a search query for a specific user
func (r *ImageRepository) SearchCount(ctx context.Context, userID int64, query string) (int, error) {
	pattern := "%" + query + "%"
	count, err := r.q.CountSearchImages(ctx, sqlc.CountSearchImagesParams{
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
		Name:   pattern,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return int(count), nil
}

// Create creates a new image for a specific user
func (r *ImageRepository) Create(ctx context.Context, image *models.Image) (*models.Image, error) {
	var description pgtype.Text
	description.String = image.Description
	description.Valid = image.Description != ""

	var userID pgtype.Int4
	if image.UserID != nil {
		userID.Int32 = int32(*image.UserID)
		userID.Valid = true
	}

	arg := sqlc.CreateImageParams{
		Name:        image.Name,
		Description: description,
		FilePath:    image.FilePath,
		MimeType:    image.MimeType,
		SizeBytes:   image.SizeBytes,
		UserID:      userID,
	}

	img, err := r.q.CreateImage(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// Update updates an image's metadata for a specific user
func (r *ImageRepository) Update(ctx context.Context, image *models.Image, userID int64) (*models.Image, error) {
	var description pgtype.Text
	description.String = image.Description
	description.Valid = image.Description != ""

	arg := sqlc.UpdateImageParams{
		ID:          int32(image.ID),
		Name:        image.Name,
		Description: description,
		UserID:      pgtype.Int4{Int32: int32(userID), Valid: true},
	}

	img, err := r.q.UpdateImage(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	return convertSQLCImage(img), nil
}

// Delete deletes an image for a specific user
func (r *ImageRepository) Delete(ctx context.Context, id int64, userID int64) error {
	err := r.q.DeleteImage(ctx, sqlc.DeleteImageParams{
		ID:     int32(id),
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// ListCursor retrieves a paginated list of images using cursor-based pagination
func (r *ImageRepository) ListCursor(ctx context.Context, userID int64, cursor int64, limit int) ([]*models.Image, error) {
	arg := sqlc.ListImagesCursorParams{
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
		ID:     int32(cursor),
		Limit:  int32(limit),
	}

	imgs, err := r.q.ListImagesCursor(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list images with cursor: %w", err)
	}

	return convertSQLCImages(imgs), nil
}

// SearchCursor searches for images using cursor-based pagination
func (r *ImageRepository) SearchCursor(ctx context.Context, userID int64, params *models.CursorSearchParams) ([]*models.Image, error) {
	pattern := "%" + params.Query + "%"
	arg := sqlc.SearchImagesCursorParams{
		UserID: pgtype.Int4{Int32: int32(userID), Valid: true},
		ID:     int32(params.Pagination.Cursor),
		Name:   pattern,
		Limit:  int32(params.Pagination.PageSize),
	}

	imgs, err := r.q.SearchImagesCursor(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to search images with cursor: %w", err)
	}

	return convertSQLCImages(imgs), nil
}

// Helper functions to convert between SQLC and domain models
func convertSQLCImage(img sqlc.Image) *models.Image {
	description := ""
	if img.Description.Valid {
		description = img.Description.String
	}

	var userID *int64
	if img.UserID.Valid {
		uid := int64(img.UserID.Int32)
		userID = &uid
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
		UserID:      userID,
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
