package auth

import (
	"net/http"

	"github.com/dmdhrumilmistry/defect-detect/pkg/config"
	"github.com/dmdhrumilmistry/defect-detect/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	store types.AuthStore
}

func NewAuthHandler(store types.AuthStore) *AuthHandler {
	handler := &AuthHandler{
		store: store,
	}

	handler.InitAuth()

	return handler
}

// InitAuth initializes the authentication providers
func (a *AuthHandler) InitAuth() {
	log.Info().Msg("Initializing Auth Providers")

	googleRedirectUrl := config.DefaultConfig.GetBaseUrl() + "/auth/google/callback"
	log.Info().Msgf("Google Redirect Url: %s", googleRedirectUrl)

	goth.UseProviders(
		google.New(
			config.DefaultConfig.GoogleClientId,
			config.DefaultConfig.GoogleClientSecret,
			googleRedirectUrl, // Redirect URL
			"email", "profile",
		),
	)
	log.Info().Msg("Initialized Auth Providers Successfully")
}

func (a *AuthHandler) RegisterRoutes(r *gin.Engine) {
	// Google auth
	r.GET("/auth/", a.GoogleAuthHandler) // GET http://domain:8080/auth/?provider=google
	r.GET("/auth/google/callback", a.GoogleCallbackHandler)

	// use authstore auth middleware if sensitive actions are
	// performed or pii details are returned in response

	log.Info().Msg("Auth Providers routes registered")
}

// AuthHandler redirects users to Google login
func (a *AuthHandler) GoogleAuthHandler(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// CallbackHandler handles Google auth callback
func (a *AuthHandler) GoogleCallbackHandler(c *gin.Context) {
	guser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Error().Err(err).Msg("failed to complete google oauth flow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	user, err := a.store.GetUserByEmail(guser.Email, config.DefaultConfig.DbQueryTimeout)
	if err != nil {
		log.Error().Err(err).Msgf("user not found for provided email id")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "ask admin to create account and provide necessary permissions"})
		return
	}

	if !user.IsActive {
		log.Warn().Msgf("inactive user tried to login: %s", user.Id)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "ask admin to reactivate account"})
		return
	}

	log.Info().Msgf("User %s logged in using google oauth", user.Id)

	// generate token
	token, err := CreateJWT(user.Id)
	if err != nil {
		log.Error().Err(err).Msgf("failed to generate token for user %s", user.Id)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token Generated Successfully",
		"token":   token,
	})
}
