package services

import (
	"os"
	"testing"
	"time"

	"bookify/internal/db"
	"bookify/internal/testutil"
)

func TestNewDriveService_Basic(t *testing.T) {
	service := NewDriveService(&db.Service{})
	if service == nil {
		t.Fatal("NewDriveService() returned nil")
	}
	// Verify it's initialized with the mock DB service
	if service.dbService == nil {
		t.Errorf("NewDriveService() didn't initialize dbService")
	}
}

func TestDriveService_TestConnection_NoAuth(t *testing.T) {
	service := &DriveService{
		dbService: &db.Service{},
	}
	account := &db.Account{
		AccessToken:  "",
		RefreshToken: "",
	}
	err := service.TestConnection(account)

	if err == nil {
		t.Errorf("TestConnection() expected error with no auth, got nil")
	}
}

func TestDriveService_TestFolderAccess_NoAuth(t *testing.T) {
	service := &DriveService{
		dbService: &db.Service{},
	}
	account := &db.Account{
		FolderID:     "test-folder-id",
		AccessToken:  "",
		RefreshToken: "",
	}
	err := service.TestFolderAccess(account)

	if err == nil {
		t.Errorf("TestFolderAccess() expected error with no auth, got nil")
	}
}

func TestDriveService_UploadFile_NoAuth(t *testing.T) {
	service := &DriveService{
		dbService: &db.Service{},
	}
	account := &db.Account{
		FolderID:     "test-folder-id",
		AccessToken:  "",
		RefreshToken: "",
	}
	filePath := testutil.CreateInvalidFile(t, "test.txt")

	url, err := service.UploadFile(account, filePath, "test.txt")

	if err == nil {
		t.Errorf("UploadFile() expected error with no auth, got nil")
	}
	if url != "" {
		t.Errorf("UploadFile() expected empty URL on error, got %v", url)
	}
}

func TestDriveService_RefreshTokenIfNeeded_NotExpired(t *testing.T) {
	service := &DriveService{
		dbService: &db.Service{},
	}
	account := &db.Account{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		TokenExpiry:  time.Now().Add(1 * time.Hour), // Not expired
	}

	err := service.RefreshTokenIfNeeded(account)
	if err != nil {
		t.Errorf("RefreshTokenIfNeeded() unexpected error: %v", err)
	}
}

func TestDriveService_RefreshTokenIfNeeded_Expired(t *testing.T) {
	// Set temporary environment variables for the test
	oldClientID := os.Getenv("GOOGLE_CLIENT_ID")
	oldClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	_ = os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	_ = os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	defer func() {
		_ = os.Setenv("GOOGLE_CLIENT_ID", oldClientID)
		_ = os.Setenv("GOOGLE_CLIENT_SECRET", oldClientSecret)
	}()

	service := NewDriveService(&db.Service{})
	account := &db.Account{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		TokenExpiry:  time.Now().Add(-1 * time.Hour), // Expired
	}

	// This will fail without a valid OAuth server, which is expected
	err := service.RefreshTokenIfNeeded(account)
	if err == nil {
		t.Errorf("RefreshTokenIfNeeded() expected error with expired token and test OAuth config")
	}
}
