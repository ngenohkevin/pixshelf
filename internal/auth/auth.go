package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ngenohkevin/pixshelf/internal/db/sqlc"
	"github.com/ngenohkevin/pixshelf/templates"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	SessionUserID = "user_id"
	SessionState  = "oauth_state"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	BaseURL            string
}

type AuthService struct {
	config *AuthConfig
	oauth  *oauth2.Config
	db     *sqlc.Queries
}

func NewAuthService(config *AuthConfig, db *sqlc.Queries) *AuthService {
	oauth := &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  config.BaseURL + "/auth/google/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	return &AuthService{
		config: config,
		oauth:  oauth,
		db:     db,
	}
}

func (a *AuthService) GetAuthURL(state string) string {
	return a.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (a *AuthService) HandleCallback(code string) (*sqlc.User, error) {
	token, err := a.oauth.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := a.oauth.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Try to get existing user
	user, err := a.db.GetUserByGoogleID(context.Background(), userInfo.ID)
	if err != nil {
		// User doesn't exist, create new one
		var avatarURL pgtype.Text
		if userInfo.Picture != "" {
			avatarURL.String = userInfo.Picture
			avatarURL.Valid = true
		}

		user, err = a.db.CreateUser(context.Background(), sqlc.CreateUserParams{
			GoogleID:  userInfo.ID,
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			AvatarUrl: avatarURL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return &user, nil
}

// RequireAuth middleware
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(SessionUserID)

		if userID == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// GetCurrentUserID gets the current user ID from context
func GetCurrentUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	// Handle different integer types that might come from session
	switch v := userID.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// GetCurrentUser gets the current user data from database
func GetCurrentUser(c *gin.Context, db *sqlc.Queries) (*sqlc.User, error) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		return nil, fmt.Errorf("no user ID in context")
	}

	user, err := db.GetUser(context.Background(), int32(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// ConvertUserToTemplateData converts sqlc.User to template UserData
func ConvertUserToTemplateData(user *sqlc.User) *templates.UserData {
	if user == nil {
		return nil
	}

	avatarURL := ""
	if user.AvatarUrl.Valid {
		avatarURL = user.AvatarUrl.String
	}

	return &templates.UserData{
		ID:        int64(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURL,
	}
}
