package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"bookify/internal/db"
)

type OAuthHandlers struct {
	db           *db.Service
	oauth2Config *oauth2.Config
	stateStore   map[string]*OAuthState
}

type OAuthState struct {
	State       string
	AccountName string
	FolderID    string
	CreatedAt   time.Time
}

func NewOAuthHandlers(dbService *db.Service) *OAuthHandlers {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		panic("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	return &OAuthHandlers{
		db: dbService,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "http://localhost:8080/oauth/callback",
			Scopes: []string{
				drive.DriveScope,
			},
			Endpoint: google.Endpoint,
		},
		stateStore: make(map[string]*OAuthState),
	}
}

func (h *OAuthHandlers) cleanupOldStates() {
	now := time.Now()
	for state, data := range h.stateStore {
		if now.Sub(data.CreatedAt) > 15*time.Minute {
			delete(h.stateStore, state)
		}
	}
}

func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *OAuthHandlers) StartOAuth(c echo.Context) error {
	h.cleanupOldStates()

	accountName := c.QueryParam("account_name")
	folderID := c.QueryParam("folder_id")

	if accountName == "" || folderID == "" {
		return c.Redirect(http.StatusFound, "/setup?error=missing_params")
	}

	state, err := generateStateToken()
	if err != nil {
		return c.Redirect(http.StatusFound, "/setup?error=state_generation_failed")
	}

	h.stateStore[state] = &OAuthState{
		State:       state,
		AccountName: accountName,
		FolderID:    folderID,
		CreatedAt:   time.Now(),
	}

	authURL := h.oauth2Config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account consent"))
	return c.Redirect(http.StatusFound, authURL)
}

func (h *OAuthHandlers) OAuthCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorMsg := c.QueryParam("error")

	if errorMsg != "" {
		return c.Redirect(http.StatusFound, "/setup?error=oauth_denied")
	}

	oauthState, exists := h.stateStore[state]
	if !exists {
		return c.Redirect(http.StatusFound, "/setup?error=invalid_state")
	}
	delete(h.stateStore, state)

	ctx := context.Background()
	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return c.Redirect(http.StatusFound, "/setup?error=token_exchange_failed")
	}

	client := h.oauth2Config.Client(ctx, token)
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return c.Redirect(http.StatusFound, "/setup?error=drive_service_failed")
	}

	_, err = driveService.Files.Get(oauthState.FolderID).
		Fields("id, name").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		c.Logger().Errorf("Folder access failed for folder ID %s: %v", oauthState.FolderID, err)
		return c.Redirect(http.StatusFound, "/setup?error=folder_access_failed")
	}

	tokenInfo, err := client.Transport.(*oauth2.Transport).Source.Token()
	if err != nil {
		return c.Redirect(http.StatusFound, "/setup?error=token_info_failed")
	}

	userEmail := ""
	if idToken, ok := tokenInfo.Extra("id_token").(string); ok {
		userEmail = h.extractEmailFromIDToken(c, idToken)
	}

	if userEmail == "" {
		aboutService, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err == nil {
			about, err := aboutService.About.Get().Fields("user").Do()
			if err == nil && about.User != nil {
				userEmail = about.User.EmailAddress
			}
		}
	}

	account := &db.Account{
		Name:         oauthState.AccountName,
		FolderID:     oauthState.FolderID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		UserEmail:    userEmail,
	}

	c.Logger().Infof("Creating account: Name=%s, FolderID=%s, Email=%s, HasAccessToken=%v, HasRefreshToken=%v",
		account.Name, account.FolderID, account.UserEmail,
		account.AccessToken != "", account.RefreshToken != "")

	if err := h.db.CreateAccountWithOAuth(account); err != nil {
		c.Logger().Errorf("Failed to create account: %v", err)
		return c.Redirect(http.StatusFound, "/setup?error=account_creation_failed")
	}

	c.Logger().Infof("Account created successfully with ID: %d", account.ID)
	return c.Redirect(http.StatusFound, "/?success=account_created")
}

func (h *OAuthHandlers) extractEmailFromIDToken(c echo.Context, idToken string) string {
	// Parse the token without verification first to extract claims
	// Note: In production, you should verify the token signature using Google's public keys
	token, _, err := jwt.NewParser().ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		c.Logger().Errorf("Failed to parse ID token: %v", err)
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Logger().Errorf("Failed to extract claims from ID token")
		return ""
	}

	// Extract email from claims
	email, ok := claims["email"].(string)
	if !ok {
		c.Logger().Errorf("Email claim not found or not a string in ID token")
		return ""
	}

	// Validate token expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			c.Logger().Errorf("ID token has expired")
			return ""
		}
	}

	// Validate issuer
	if iss, ok := claims["iss"].(string); ok {
		if iss != "https://accounts.google.com" && iss != "accounts.google.com" {
			c.Logger().Errorf("Invalid issuer in ID token: %s", iss)
			return ""
		}
	}

	// Validate audience (should match your client ID)
	if aud, ok := claims["aud"].(string); ok {
		if aud != h.oauth2Config.ClientID {
			c.Logger().Errorf("Invalid audience in ID token: expected %s, got %s", h.oauth2Config.ClientID, aud)
			return ""
		}
	}

	return email
}
