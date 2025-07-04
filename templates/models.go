package templates

import (
	"time"
)

// ImageData represents the image data model for templates
type ImageData struct {
	ID          int64
	Name        string
	Description string
	URL         string
	PublicURL   string
	MimeType    string
	SizeBytes   int64
	CreatedAt   time.Time
}

// Pagination represents pagination data for templates
type Pagination struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	HasPrev     bool
	HasNext     bool
	Query       string
}

// UserData represents the user data model for templates
type UserData struct {
	ID        int64
	Name      string
	Email     string
	AvatarURL string
}
