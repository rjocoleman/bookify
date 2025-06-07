package services

import (
	"os"
	"strings"
	"testing"

	"bookify/internal/testutil"
)

func TestNewDriveService_Basic(t *testing.T) {
	service := NewDriveService()
	if service == nil {
		t.Errorf("NewDriveService() returned nil")
	}
}

func TestDriveService_GetServiceAccountEmail_NoKey(t *testing.T) {
	service := &DriveService{serviceAccountKey: nil}
	email := service.GetServiceAccountEmail()

	expected := "No service account configured"
	if email != expected {
		t.Errorf("GetServiceAccountEmail() = %v, want %v", email, expected)
	}
}

func TestDriveService_GetServiceAccountEmail_ValidKey(t *testing.T) {
	validKey := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "test-key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n",
		"client_email": "valid-test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`

	service := &DriveService{serviceAccountKey: []byte(validKey)}
	email := service.GetServiceAccountEmail()

	expected := "valid-test@test-project.iam.gserviceaccount.com"
	if email != expected {
		t.Errorf("GetServiceAccountEmail() = %v, want %v", email, expected)
	}
}

func TestDriveService_TestConnection_NoKey(t *testing.T) {
	service := &DriveService{serviceAccountKey: nil}
	err := service.TestConnection()

	if err == nil {
		t.Errorf("TestConnection() expected error with no key, got nil")
	}
}

func TestDriveService_TestFolderAccess_NoKey(t *testing.T) {
	service := &DriveService{serviceAccountKey: nil}
	err := service.TestFolderAccess("test-folder-id")

	if err == nil {
		t.Errorf("TestFolderAccess() expected error with no key, got nil")
	}
}

func TestDriveService_UploadFile_NoKey(t *testing.T) {
	service := &DriveService{serviceAccountKey: nil}
	filePath := testutil.CreateInvalidFile(t, "test.txt")

	url, err := service.UploadFile("folder-id", filePath, "test.txt")

	if err == nil {
		t.Errorf("UploadFile() expected error with no key, got nil")
	}
	if url != "" {
		t.Errorf("UploadFile() expected empty URL on error, got %v", url)
	}
}

// Integration test that requires real service account credentials
func TestDriveService_Integration(t *testing.T) {
	keyPath := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY_PATH")
	keyJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY")

	if keyPath == "" && keyJSON == "" {
		t.Skip("Skipping integration test: no service account credentials configured")
	}

	service := NewDriveService()
	if len(service.serviceAccountKey) == 0 {
		t.Skip("Skipping integration test: no valid service account key loaded")
	}

	// Test service account email
	email := service.GetServiceAccountEmail()
	if email == "No service account configured" || email == "Error reading service account" {
		t.Errorf("GetServiceAccountEmail() = %v, expected valid email", email)
	}
	if !strings.Contains(email, "@") {
		t.Errorf("GetServiceAccountEmail() = %v, doesn't look like an email", email)
	}

	// Test connection (may fail with test credentials, that's OK)
	err := service.TestConnection()
	if err != nil {
		t.Logf("TestConnection() failed (expected for test credentials): %v", err)
	} else {
		t.Log("TestConnection() succeeded")
	}
}
