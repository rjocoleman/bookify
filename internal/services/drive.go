package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"bookify/internal/db"
)

type DriveService struct {
	dbService    *db.Service
	oauth2Config *oauth2.Config
}

func NewDriveService(dbService *db.Service) *DriveService {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Println("WARNING: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set for OAuth")
	}

	return &DriveService{
		dbService: dbService,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "http://localhost:8080/oauth/callback",
			Scopes: []string{
				drive.DriveScope,
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (d *DriveService) getOAuthClient(account *db.Account) (*drive.Service, error) {
	if account.AccessToken == "" || account.RefreshToken == "" {
		return nil, fmt.Errorf("account not authenticated with OAuth")
	}

	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       account.TokenExpiry,
		TokenType:    "Bearer",
	}

	tokenSource := d.oauth2Config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if newToken.AccessToken != token.AccessToken {
		account.AccessToken = newToken.AccessToken
		account.TokenExpiry = newToken.Expiry
		if err := d.dbService.UpdateAccount(account); err != nil {
			log.Printf("Failed to update refreshed token: %v", err)
		}
	}

	client := d.oauth2Config.Client(context.Background(), newToken)
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return service, nil
}

func (d *DriveService) UploadFile(account *db.Account, filePath, fileName string) (string, error) {
	service, err := d.getOAuthClient(account)
	if err != nil {
		return "", err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close file: %v", closeErr)
		}
	}()

	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{account.FolderID},
	}

	res, err := service.Files.Create(driveFile).
		Media(file).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	shareURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", res.Id)
	return shareURL, nil
}

func (d *DriveService) TestConnection(account *db.Account) error {
	service, err := d.getOAuthClient(account)
	if err != nil {
		return err
	}

	_, err = service.About.Get().Fields("user").Do()
	if err != nil {
		return fmt.Errorf("failed to test Drive connection: %w", err)
	}

	return nil
}

func (d *DriveService) TestFolderAccess(account *db.Account) error {
	service, err := d.getOAuthClient(account)
	if err != nil {
		return err
	}

	log.Printf("Testing access to folder ID: %s", account.FolderID)

	file, err := service.Files.Get(account.FolderID).
		Fields("id, name, mimeType, owners").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		log.Printf("Google API error details: %v", err)
		return fmt.Errorf("failed to access folder: %w", err)
	}

	log.Printf("✓ Successfully accessed folder: %s (ID: %s, Type: %s)", file.Name, file.Id, file.MimeType)
	return nil
}

func (d *DriveService) RefreshTokenIfNeeded(account *db.Account) error {
	if time.Now().After(account.TokenExpiry) {
		_, err := d.getOAuthClient(account)
		return err
	}
	return nil
}
