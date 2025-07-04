package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngenohkevin/pixshelf/internal/auth"
	"github.com/ngenohkevin/pixshelf/internal/db/sqlc"
	"github.com/ngenohkevin/pixshelf/internal/service"
	"github.com/ngenohkevin/pixshelf/templates"
)

// UIHandler handles UI requests
type UIHandler struct {
	service *service.ImageService
	db      *sqlc.Queries
}

// NewUIHandler creates a new UIHandler
func NewUIHandler(service *service.ImageService, db *sqlc.Queries) *UIHandler {
	return &UIHandler{
		service: service,
		db:      db,
	}
}

// Home renders the homepage
func (h *UIHandler) Home(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// Get current user data
	sqlcUser, err := auth.GetCurrentUser(c, h.db)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	user := auth.ConvertUserToTemplateData(sqlcUser)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get search query if present
	query := c.Query("q")

	var images []*templates.ImageData
	var pagination *templates.Pagination

	if query != "" {
		// Perform search
		imgs, p, err := h.service.Search(c.Request.Context(), userID, query, page, pageSize)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		// Convert to template models
		images = make([]*templates.ImageData, len(imgs))
		for i, img := range imgs {
			images[i] = &templates.ImageData{
				ID:          img.ID,
				Name:        img.Name,
				Description: img.Description,
				URL:         img.URL,
				PublicURL:   img.PublicURL,
				MimeType:    img.MimeType,
				SizeBytes:   img.SizeBytes,
				CreatedAt:   img.CreatedAt,
			}
		}

		pagination = &templates.Pagination{
			CurrentPage: p.Page,
			TotalPages:  (p.Total + p.PageSize - 1) / p.PageSize,
			TotalItems:  p.Total,
			HasPrev:     p.Page > 1,
			HasNext:     p.Page*p.PageSize < p.Total,
			Query:       query,
		}
	} else {
		// List all images
		imgs, p, err := h.service.List(c.Request.Context(), userID, page, pageSize)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		// Convert to template models
		images = make([]*templates.ImageData, len(imgs))
		for i, img := range imgs {
			images[i] = &templates.ImageData{
				ID:          img.ID,
				Name:        img.Name,
				Description: img.Description,
				URL:         img.URL,
				PublicURL:   img.PublicURL,
				MimeType:    img.MimeType,
				SizeBytes:   img.SizeBytes,
				CreatedAt:   img.CreatedAt,
			}
		}

		pagination = &templates.Pagination{
			CurrentPage: p.Page,
			TotalPages:  (p.Total + p.PageSize - 1) / p.PageSize,
			TotalItems:  p.Total,
			HasPrev:     p.Page > 1,
			HasNext:     p.Page*p.PageSize < p.Total,
		}
	}

	component := templates.Home(images, pagination, query, user)
	component.Render(c.Request.Context(), c.Writer)
}

// ImageDetail renders the image detail page
func (h *UIHandler) ImageDetail(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// Get current user data
	sqlcUser, err := auth.GetCurrentUser(c, h.db)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	user := auth.ConvertUserToTemplateData(sqlcUser)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	img, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

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

	component := templates.ImageDetail(imageData, user)
	component.Render(c.Request.Context(), c.Writer)
}

// Upload renders the upload page
func (h *UIHandler) Upload(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// Get current user data
	sqlcUser, err := auth.GetCurrentUser(c, h.db)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	user := auth.ConvertUserToTemplateData(sqlcUser)

	component := templates.Upload(user)
	component.Render(c.Request.Context(), c.Writer)
}

// Edit renders the edit page
func (h *UIHandler) Edit(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// Get current user data
	sqlcUser, err := auth.GetCurrentUser(c, h.db)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	user := auth.ConvertUserToTemplateData(sqlcUser)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	img, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	imageData := &templates.ImageData{
		ID:          img.ID,
		Name:        img.Name,
		Description: img.Description,
		URL:         img.URL,
		PublicURL:   img.PublicURL,
	}

	component := templates.Edit(imageData, user)
	component.Render(c.Request.Context(), c.Writer)
}

// SearchResults renders the search results for HTMX requests
func (h *UIHandler) SearchResults(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	query := c.Query("q")
	if query == "" {
		c.Status(http.StatusBadRequest)
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

	imgs, p, err := h.service.Search(c.Request.Context(), userID, query, page, pageSize)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Convert to template models
	images := make([]*templates.ImageData, len(imgs))
	for i, img := range imgs {
		images[i] = &templates.ImageData{
			ID:          img.ID,
			Name:        img.Name,
			Description: img.Description,
			URL:         img.URL,
			PublicURL:   img.PublicURL,
			MimeType:    img.MimeType,
			SizeBytes:   img.SizeBytes,
			CreatedAt:   img.CreatedAt,
		}
	}

	pagination := &templates.Pagination{
		CurrentPage: p.Page,
		TotalPages:  (p.Total + p.PageSize - 1) / p.PageSize,
		TotalItems:  p.Total,
		HasPrev:     p.Page > 1,
		HasNext:     p.Page*p.PageSize < p.Total,
		Query:       query,
	}

	component := templates.ImageList(images, pagination)
	component.Render(c.Request.Context(), c.Writer)
}

// RegisterRoutes registers the UI routes
func (h *UIHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/", h.Home)
	router.GET("/view-image/:id", h.ImageDetail) // Changed from /images/:id to avoid conflict
	router.GET("/upload", h.Upload)
	router.GET("/view-image/:id/edit", h.Edit) // Changed from /images/:id/edit to avoid conflict
	router.GET("/search", h.SearchResults)
}
