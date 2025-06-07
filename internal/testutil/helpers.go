package testutil

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}

// CreateTestEPUB creates a minimal valid EPUB file for testing
func CreateTestEPUB(t *testing.T, filename string) string {
	t.Helper()

	tempDir := t.TempDir()
	epubPath := filepath.Join(tempDir, filename)

	// Create a minimal EPUB structure (just a ZIP with mimetype)
	file, err := os.Create(epubPath)
	if err != nil {
		t.Fatalf("Failed to create test EPUB: %v", err)
	}
	defer func() {
		_ = file.Close() // Error ignored in test helper
	}()

	// Write minimal ZIP header for testing
	_, err = file.Write([]byte("PK\x03\x04"))
	if err != nil {
		t.Fatalf("Failed to write EPUB header: %v", err)
	}

	return epubPath
}

// CreateInvalidFile creates a file that's not a valid EPUB
func CreateInvalidFile(t *testing.T, filename string) string {
	t.Helper()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, []byte("This is not an EPUB file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	return filePath
}

// CreateMultipartRequest creates a multipart form request for file upload testing
func CreateMultipartRequest(t *testing.T, method, url string, files map[string]string, fields map[string]string) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		if err != nil {
			t.Fatalf("Failed to write form field: %v", err)
		}
	}

	// Add files
	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("Failed to open file %s: %v", filePath, err)
		}
		defer func() {
			_ = file.Close() // Error ignored in test helper
		}()

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			t.Fatalf("Failed to copy file content: %v", err)
		}
	}

	err := writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// AssertResponseStatus checks if the response has the expected status code
func AssertResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()

	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response body: %s",
			expectedStatus, rec.Code, rec.Body.String())
	}
}

// AssertResponseContains checks if the response body contains the expected string
func AssertResponseContains(t *testing.T, rec *httptest.ResponseRecorder, expectedContent string) {
	t.Helper()

	body := rec.Body.String()
	if !strings.Contains(body, expectedContent) {
		t.Errorf("Expected response to contain %q, got: %s", expectedContent, body)
	}
}

// SetTestEnv sets environment variables for testing and returns a cleanup function
func SetTestEnv(t *testing.T, env map[string]string) func() {
	t.Helper()

	originalValues := make(map[string]string)

	for key, value := range env {
		originalValues[key] = os.Getenv(key)
		if err := os.Setenv(key, value); err != nil {
			t.Errorf("Failed to set env var %s: %v", key, err)
		}
	}

	return func() {
		for key, originalValue := range originalValues {
			if originalValue == "" {
				if err := os.Unsetenv(key); err != nil {
					t.Errorf("Failed to unset env var %s: %v", key, err)
				}
			} else {
				if err := os.Setenv(key, originalValue); err != nil {
					t.Errorf("Failed to restore env var %s: %v", key, err)
				}
			}
		}
	}
}

// CreateTestServiceAccount creates a mock service account JSON for testing
func CreateTestServiceAccount(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test-service-account.json")

	serviceAccountJSON := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "test-key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`

	err := os.WriteFile(keyPath, []byte(serviceAccountJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service account file: %v", err)
	}

	return keyPath
}
