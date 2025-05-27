package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ngenohkevin/pixshelf/templates"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.GET("/login", h.Login)
		auth.GET("/google", h.GoogleAuth)
		auth.GET("/google/callback", h.GoogleCallback)
		auth.POST("/logout", h.Logout)
	}

	r.GET("/login", h.ShowLogin)
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	// Check if user is already logged in
	session := sessions.Default(c)
	if userID := session.Get(SessionUserID); userID != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	component := templates.Login()
	component.Render(c.Request.Context(), c.Writer)
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/auth/google")
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	// Generate state string
	state := generateStateString()

	session := sessions.Default(c)
	session.Set(SessionState, state)
	session.Save()

	authURL := h.authService.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	session := sessions.Default(c)
	savedState := session.Get(SessionState)

	// Verify state
	if savedState == nil || savedState != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Clear state from session
	session.Delete(SessionState)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
		return
	}

	user, err := h.authService.HandleCallback(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Save user ID in session - make sure it's saved properly
	session.Set(SessionUserID, user.ID)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
		return
	}

	// Handle both GET and POST requests for logout
	if c.Request.Method == "POST" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusOK)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func generateStateString() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
