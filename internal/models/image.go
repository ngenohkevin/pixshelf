package models

import (
	"time"
)

// Image represents an image in the system
type Image struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	FilePath    string    `json:"file_path"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ImageURL returns the URL for accessing the image
func (i *Image) ImageURL(baseURL string) string {
	// For backward compatibility, keep the original URL structure
	return baseURL + "/static/images/" + i.FilePath
}

// PublicImageURL returns the public URL for accessing the image in a more shareable format
func (i *Image) PublicImageURL(baseURL string) string {
	// Format: https://pixshelf.perigrine.cloud/public-images/{uuid}_{image_name}.extension
	return baseURL + "/public-images/" + i.FilePath
}

// PublicImage represents the public-facing image data
type PublicImage struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	PublicURL   string    `json:"public_url"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewPublicImage converts an Image to a PublicImage
func NewPublicImage(image *Image, baseURL string) *PublicImage {
	return &PublicImage{
		ID:          image.ID,
		Name:        image.Name,
		Description: image.Description,
		URL:         image.ImageURL(baseURL),
		PublicURL:   image.PublicImageURL(baseURL),
		MimeType:    image.MimeType,
		SizeBytes:   image.SizeBytes,
		CreatedAt:   image.CreatedAt,
		UpdatedAt:   image.UpdatedAt,
	}
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// SearchParams represents search parameters
type SearchParams struct {
	Query      string `json:"query"`
	Pagination *Pagination
}
