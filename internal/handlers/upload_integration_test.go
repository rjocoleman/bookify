package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bookify/internal/db"
	"bookify/internal/services"
	"bookify/internal/testutil"

	"github.com/labstack/echo/v4"
)

func TestUploadHandler_MultipleEPUBs(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := services.NewDriveService(dbService)

	// Use t.TempDir() for test temp directory
	tempDir := t.TempDir()

	handlers := &Handlers{
		DB:      dbService,
		Drive:   driveService,
		TempDir: tempDir,
	}

	// Create a test account
	account, err := dbService.CreateAccount("test-account", "folder-123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	// Create Echo instance
	e := echo.New()

	// Create multipart form with multiple files
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add account_id field
	err = writer.WriteField("account_id", fmt.Sprintf("%d", account.ID))
	if err != nil {
		t.Fatalf("Failed to write account_id field: %v", err)
	}

	// Create mock EPUB files
	epubFiles := []string{"book1.epub", "book2.epub", "book3.epub"}
	for _, filename := range epubFiles {
		part, err := writer.CreateFormFile("files", filename)
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		// Write mock EPUB content (minimal valid EPUB structure)
		mockContent := "PK" + strings.Repeat("\x00", 100) // Simulate ZIP file header
		_, err = io.WriteString(part, mockContent)
		if err != nil {
			t.Fatalf("Failed to write mock content: %v", err)
		}
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err = handlers.UploadHandler(c)
	if err != nil {
		t.Fatalf("UploadHandler() error = %v", err)
	}

	// Check response status
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify all jobs were created
	jobs, err := dbService.ListRecentJobs(10)
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs to be created, got %d", len(jobs))
	}

	// Verify all jobs have 'queued' status
	for i, job := range jobs {
		if job.Status != "queued" {
			t.Errorf("Job %d: expected status 'queued', got '%s'", i, job.Status)
		}
		if job.AccountID != account.ID {
			t.Errorf("Job %d: expected AccountID %d, got %d", i, account.ID, job.AccountID)
		}
	}

	// Optional: Test that jobs can be processed in order
	// This demonstrates that the queue worker would process them correctly
	for i := 0; i < 3; i++ {
		nextJob, err := dbService.GetNextQueuedJob()
		if err != nil {
			t.Fatalf("GetNextQueuedJob() error = %v", err)
		}
		if nextJob == nil {
			t.Fatalf("Expected to get queued job %d, got nil", i)
		}

		// Verify it's one of our uploaded files
		found := false
		for _, filename := range epubFiles {
			if nextJob.OriginalFilename == filename {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected job filename: %s", nextJob.OriginalFilename)
		}

		// Simulate processing
		err = dbService.MarkJobCompleted(nextJob.ID, "processed-"+nextJob.OriginalFilename, "http://drive.example.com/file")
		if err != nil {
			t.Errorf("Failed to mark job completed: %v", err)
		}
	}

	// Verify no more queued jobs
	nextJob, err := dbService.GetNextQueuedJob()
	if err != nil {
		t.Errorf("GetNextQueuedJob() error = %v", err)
	}
	if nextJob != nil {
		t.Errorf("Expected no more queued jobs, but found job %s", nextJob.ID)
	}
}

func TestUploadHandler_MultipleEPUBs_ProcessingOrder(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := services.NewDriveService(dbService)

	// Use t.TempDir() for test temp directory
	tempDir := t.TempDir()

	handlers := &Handlers{
		DB:      dbService,
		Drive:   driveService,
		TempDir: tempDir,
	}

	// Create a test account
	account, err := dbService.CreateAccount("test-account", "folder-123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	// Create Echo instance
	e := echo.New()

	// Upload files with slight delays to ensure order
	uploadedFiles := []string{}
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("book%d.epub", i)
		uploadedFiles = append(uploadedFiles, filename)

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		err = writer.WriteField("account_id", fmt.Sprintf("%d", account.ID))
		if err != nil {
			t.Fatalf("Failed to write account_id field: %v", err)
		}

		part, err := writer.CreateFormFile("files", filename)
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		// Write mock EPUB content
		mockContent := "PK" + strings.Repeat("\x00", 100)
		_, err = io.WriteString(part, mockContent)
		if err != nil {
			t.Fatalf("Failed to write mock content: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed to close multipart writer: %v", err)
		}

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call the handler
		err = handlers.UploadHandler(c)
		if err != nil {
			t.Fatalf("UploadHandler() error = %v", err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Process all jobs and verify FIFO order
	processedFiles := []string{}
	for i := 0; i < 5; i++ {
		job, err := dbService.GetNextQueuedJob()
		if err != nil {
			t.Fatalf("GetNextQueuedJob() error = %v", err)
		}
		if job == nil {
			t.Fatalf("Expected job %d but got nil", i)
		}

		processedFiles = append(processedFiles, job.OriginalFilename)

		// Mark as completed
		err = dbService.MarkJobCompleted(job.ID, "processed-"+job.OriginalFilename, "http://drive.example.com/file")
		if err != nil {
			t.Errorf("Failed to mark job completed: %v", err)
		}
	}

	// Verify FIFO order
	for i := 0; i < 5; i++ {
		if processedFiles[i] != uploadedFiles[i] {
			t.Errorf("Processing order violation: expected %s at position %d, got %s",
				uploadedFiles[i], i, processedFiles[i])
		}
	}
}
