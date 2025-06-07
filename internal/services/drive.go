package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveService struct {
	serviceAccountKey []byte
}

func NewDriveService() *DriveService {
	// Try to load service account key from file path first
	keyPath := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY_PATH")
	if keyPath != "" {
		log.Printf("Loading service account key from file: %s", keyPath)
		keyData, err := os.ReadFile(keyPath)
		if err == nil {
			log.Println("✓ Service account key loaded from file")
			return &DriveService{
				serviceAccountKey: keyData,
			}
		}
		log.Printf("Failed to read service account key file: %v", err)
	}

	// Fall back to reading key directly from environment variable
	keyJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY")
	if keyJSON != "" {
		log.Println("Loading service account key from environment variable")
		return &DriveService{
			serviceAccountKey: []byte(keyJSON),
		}
	}

	// No service account key found
	log.Println("WARNING: No service account key found. Set GOOGLE_SERVICE_ACCOUNT_KEY_PATH or GOOGLE_SERVICE_ACCOUNT_KEY")
	return &DriveService{}
}

func (d *DriveService) getClient() (*drive.Service, error) {
	if len(d.serviceAccountKey) == 0 {
		return nil, fmt.Errorf("no service account key configured")
	}

	config, err := google.JWTConfigFromJSON(d.serviceAccountKey, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %w", err)
	}

	log.Printf("Service account email: %s", config.Email)

	client := config.Client(context.Background())
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return service, nil
}

func (d *DriveService) UploadFile(folderID, filePath, fileName string) (string, error) {
	service, err := d.getClient()
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
		Parents: []string{folderID},
	}

	res, err := service.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	shareURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", res.Id)
	return shareURL, nil
}

func (d *DriveService) TestConnection() error {
	service, err := d.getClient()
	if err != nil {
		return err
	}

	_, err = service.About.Get().Fields("user").Do()
	if err != nil {
		return fmt.Errorf("failed to test Drive connection: %w", err)
	}

	return nil
}

// TestFolderAccess verifies the service account has access to a specific folder
func (d *DriveService) TestFolderAccess(folderID string) error {
	service, err := d.getClient()
	if err != nil {
		return err
	}

	log.Printf("Testing access to folder ID: %s", folderID)

	file, err := service.Files.Get(folderID).Fields("id, name, mimeType, owners").Do()
	if err != nil {
		log.Printf("Google API error details: %v", err)
		return fmt.Errorf("failed to access folder: %w", err)
	}

	log.Printf("✓ Successfully accessed folder: %s (ID: %s, Type: %s)", file.Name, file.Id, file.MimeType)
	return nil
}

// GetServiceAccountEmail returns the email address of the configured service account
func (d *DriveService) GetServiceAccountEmail() string {
	if len(d.serviceAccountKey) == 0 {
		return "No service account configured"
	}

	config, err := google.JWTConfigFromJSON(d.serviceAccountKey, drive.DriveScope)
	if err != nil {
		return "Error reading service account"
	}

	return config.Email
}
