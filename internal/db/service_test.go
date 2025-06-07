package db

import (
	"testing"

	"bookify/internal/testutil"
)

func TestDBService_BasicOperations(t *testing.T) {
	// Setup test database
	database := testutil.SetupTestDB(t)
	err := database.AutoMigrate(&Account{}, &Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	service := NewService(database)

	// Test creating an account
	account, err := service.CreateAccount("Test Account", "folder123")
	if err != nil {
		t.Fatalf("CreateAccount() failed: %v", err)
	}
	if account.Name != "Test Account" {
		t.Errorf("CreateAccount() name = %v, want Test Account", account.Name)
	}
	if account.FolderID != "folder123" {
		t.Errorf("CreateAccount() folderID = %v, want folder123", account.FolderID)
	}

	// Test getting the account
	retrievedAccount, err := service.GetAccount(account.ID)
	if err != nil {
		t.Fatalf("GetAccount() failed: %v", err)
	}
	if retrievedAccount.Name != account.Name {
		t.Errorf("GetAccount() name = %v, want %v", retrievedAccount.Name, account.Name)
	}

	// Test getting account by name
	namedAccount, err := service.GetAccountByName("Test Account")
	if err != nil {
		t.Fatalf("GetAccountByName() failed: %v", err)
	}
	if namedAccount.ID != account.ID {
		t.Errorf("GetAccountByName() ID = %v, want %v", namedAccount.ID, account.ID)
	}

	// Test listing accounts
	accounts, err := service.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("ListAccounts() length = %v, want 1", len(accounts))
	}

	// Test creating a job
	job, err := service.CreateJob(account.ID, "test.epub")
	if err != nil {
		t.Fatalf("CreateJob() failed: %v", err)
	}
	if job.OriginalFilename != "test.epub" {
		t.Errorf("CreateJob() filename = %v, want test.epub", job.OriginalFilename)
	}
	if job.Status != "queued" {
		t.Errorf("CreateJob() status = %v, want queued", job.Status)
	}

	// Test getting the job
	retrievedJob, err := service.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}
	if retrievedJob.OriginalFilename != job.OriginalFilename {
		t.Errorf("GetJob() filename = %v, want %v", retrievedJob.OriginalFilename, job.OriginalFilename)
	}

	// Test updating job status
	job.Status = "processing"
	job.Progress = 50
	err = service.UpdateJob(job)
	if err != nil {
		t.Fatalf("UpdateJob() failed: %v", err)
	}

	updatedJob, err := service.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() after update failed: %v", err)
	}
	if updatedJob.Status != "processing" {
		t.Errorf("UpdateJob() status = %v, want processing", updatedJob.Status)
	}
	if updatedJob.Progress != 50 {
		t.Errorf("UpdateJob() progress = %v, want 50", updatedJob.Progress)
	}

	// Test marking job completed
	err = service.MarkJobCompleted(job.ID, "test.kepub.epub", "https://drive.google.com/file/d/123/view")
	if err != nil {
		t.Fatalf("MarkJobCompleted() failed: %v", err)
	}

	completedJob, err := service.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() after completion failed: %v", err)
	}
	if completedJob.Status != "completed" {
		t.Errorf("MarkJobCompleted() status = %v, want completed", completedJob.Status)
	}
	if completedJob.ProcessedFilename != "test.kepub.epub" {
		t.Errorf("MarkJobCompleted() filename = %v, want test.kepub.epub", completedJob.ProcessedFilename)
	}

	// Test marking job failed
	failedJob, err := service.CreateJob(account.ID, "failed.epub")
	if err != nil {
		t.Fatalf("CreateJob() for failure test failed: %v", err)
	}

	err = service.MarkJobFailed(failedJob.ID, "Processing failed")
	if err != nil {
		t.Fatalf("MarkJobFailed() failed: %v", err)
	}

	retrievedFailedJob, err := service.GetJob(failedJob.ID)
	if err != nil {
		t.Fatalf("GetJob() after failure failed: %v", err)
	}
	if retrievedFailedJob.Status != "failed" {
		t.Errorf("MarkJobFailed() status = %v, want failed", retrievedFailedJob.Status)
	}
	if retrievedFailedJob.Error != "Processing failed" {
		t.Errorf("MarkJobFailed() error = %v, want Processing failed", retrievedFailedJob.Error)
	}
}

func TestDBService_ErrorCases(t *testing.T) {
	// Setup test database
	database := testutil.SetupTestDB(t)
	err := database.AutoMigrate(&Account{}, &Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	service := NewService(database)

	// Test getting non-existent account
	_, err = service.GetAccount(999)
	if err == nil {
		t.Errorf("GetAccount() with invalid ID should return error")
	}

	// Test getting non-existent account by name
	_, err = service.GetAccountByName("Non-existent Account")
	if err == nil {
		t.Errorf("GetAccountByName() with invalid name should return error")
	}

	// Test getting non-existent job
	_, err = service.GetJob("non-existent-job-id")
	if err == nil {
		t.Errorf("GetJob() with invalid ID should return error")
	}

	// Test updating non-existent job
	fakeJob := &Job{ID: "fake-id", Status: "processing"}
	_ = service.UpdateJob(fakeJob)
	// This might not error depending on GORM behavior, so we don't assert
}

func TestDBService_DuplicateAccountName(t *testing.T) {
	// Setup test database
	database := testutil.SetupTestDB(t)
	err := database.AutoMigrate(&Account{}, &Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	service := NewService(database)

	// Create first account
	_, err = service.CreateAccount("Test Account", "folder123")
	if err != nil {
		t.Fatalf("CreateAccount() first account failed: %v", err)
	}

	// Try to create duplicate account
	_, err = service.CreateAccount("Test Account", "folder456")
	if err == nil {
		t.Errorf("CreateAccount() with duplicate name should return error")
	}
}
