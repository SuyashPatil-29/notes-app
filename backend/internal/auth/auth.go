package auth

import (
	"backend/db"
	"backend/internal/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"
)

const (
	MaxAge = 86400 * 30
)

var Store *sessions.CookieStore

func InitAuth() {
	var isProd bool
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("Error loading .env file, using environment variables")
	}

	if os.Getenv("IsProd") == "true" {
		isProd = true
	} else {
		isProd = false
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	sessionSecret := os.Getenv("SESSION_SECRET")

	if callbackURL == "" {
		callbackURL = "http://localhost:8080/auth/google/callback"
	}

	if sessionSecret == "" {
		sessionSecret = "MyRandomSecretKeyPleaseChangeThis32"
	}

	// Make sure the key is at least 32 bytes
	key := []byte(sessionSecret)
	if len(key) < 32 {
		// Pad the key to 32 bytes
		key = append(key, make([]byte, 32-len(key))...)
	}

	Store = sessions.NewCookieStore(key)
	Store.MaxAge(MaxAge)

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}

	gothic.Store = Store

	goth.UseProviders(google.New(googleClientId, googleClientSecret, callbackURL))
}

// BeginAuth starts the authentication process
func BeginAuth(c *gin.Context) {
	provider := c.Param("provider")

	// Set provider in query if not already set
	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// AuthCallback handles the OAuth callback
func AuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Print("Error completing auth: ", err)
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=auth_failed")
		return
	}

	// Check if user exists in database
	var user models.User
	result := db.DB.Where("email = ?", gothUser.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		imageUrl := gothUser.AvatarURL
		user = models.User{
			Name:     gothUser.Name,
			Email:    gothUser.Email,
			ImageUrl: &imageUrl,
		}

		if err := db.DB.Create(&user).Error; err != nil {
			log.Print("Error creating user in db: ", err)
			c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=db_error")
			return
		}
	}

	// Store user in session
	session, _ := Store.Get(c.Request, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Values["name"] = user.Name
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	// Redirect to frontend with user info
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?login=success")
}

// GetCurrentUser returns the currently logged-in user
func GetCurrentUser(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Fetch user from database
	var user models.AuthenticatedUser
	if err := db.DB.First(&user, userID).Error; err != nil {
		log.Print("User not found in database: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout logs out the user
func Logout(c *gin.Context) {
	provider := c.Param("provider")

	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	// Clear session
	session, _ := Store.Get(c.Request, "auth-session")
	session.Values["user_id"] = nil
	session.Values["email"] = nil
	session.Values["name"] = nil
	session.Options.MaxAge = -1
	session.Save(c.Request, c.Writer)

	if err := gothic.Logout(c.Writer, c.Request); err != nil {
		log.Print("Error during logout: ", err)
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=logout_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?logout=success")
}
