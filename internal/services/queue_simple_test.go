package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bookify/internal/db"
	"bookify/internal/testutil"
)

func TestSimpleQueueService(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := NewDriveService(dbService)

	queue := NewQueueService(dbService, driveService)
	if queue == nil {
		t.Fatalf("NewQueueService() returned nil")
	}

	// Test that queue service has correct properties
	if queue.tempDir != "./temp" {
		t.Errorf("Expected tempDir to be './temp', got %v", queue.tempDir)
	}

	if queue.processor == nil {
		t.Errorf("Expected processor to be initialized")
	}
}

func TestQueueService_StartStop(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := NewDriveService(dbService)
	queue := NewQueueService(dbService, driveService)

	// Start the worker in a goroutine
	go queue.StartWorker()

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the worker
	queue.Stop()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)

	// Test passes if no panic occurs
}

func TestQueueService_CleanupOldFiles(t *testing.T) {
	tempDir := t.TempDir()

	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := NewDriveService(dbService)

	queue := &QueueService{
		db:        dbService,
		drive:     driveService,
		processor: NewProcessorService(),
		tempDir:   tempDir,
		stopCh:    make(chan bool),
	}

	// Create test files with different ages
	now := time.Now()
	oldFile := filepath.Join(tempDir, "old-file.txt")
	newFile := filepath.Join(tempDir, "new-file.txt")

	// Create files
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// Modify the old file's timestamp to be more than 1 hour old
	oldTime := now.Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}

	// Run cleanup
	queue.cleanupOldFiles()

	// Check that old file was deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("cleanupOldFiles() should have deleted old file")
	}

	// Check that new file still exists
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Errorf("cleanupOldFiles() should not have deleted new file")
	}
}

func TestQueueService_ProcessNextJob_NoJobs(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := NewDriveService(dbService)
	queue := NewQueueService(dbService, driveService)

	// Should not panic when no jobs exist
	queue.processNextJob()

	// Test passes if no panic occurs
}
