package services

import (
	"testing"
	"time"

	"bookify/internal/db"
	"bookify/internal/testutil"
)

func TestQueueService_ProcessMultipleJobs(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)

	// Create a test account
	account, err := dbService.CreateAccount("test-account", "folder-123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	// Create 5 jobs with different timestamps
	jobIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		job, err := dbService.CreateJob(account.ID, "test"+string(rune(i))+"file.epub")
		if err != nil {
			t.Fatalf("Failed to create job %d: %v", i, err)
		}
		jobIDs[i] = job.ID

		// Sleep briefly to ensure different created_at timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Process jobs by fetching them in order and verify FIFO order
	processedOrder := make([]string, 0)

	// Process all jobs using GetNextQueuedJob
	for i := 0; i < 5; i++ {
		job, err := dbService.GetNextQueuedJob()
		if err != nil {
			t.Fatalf("GetNextQueuedJob() error = %v", err)
		}
		if job == nil {
			t.Fatalf("Expected job %d but got nil", i)
		}

		processedOrder = append(processedOrder, job.ID)

		// Simulate processing by marking as completed
		err = dbService.MarkJobCompleted(job.ID, "processed-"+job.OriginalFilename, "http://drive.example.com/file")
		if err != nil {
			t.Errorf("Failed to mark job completed: %v", err)
		}
	}

	// Verify all jobs were processed in FIFO order
	if len(processedOrder) != 5 {
		t.Errorf("Expected 5 jobs to be processed, got %d", len(processedOrder))
	}

	for i, jobID := range processedOrder {
		if jobID != jobIDs[i] {
			t.Errorf("Expected job %d to be %s, got %s (FIFO order violation)", i, jobIDs[i], jobID)
		}
	}

	// Verify no more jobs to process
	job, err := dbService.GetNextQueuedJob()
	if err != nil {
		t.Errorf("GetNextQueuedJob() error = %v", err)
	}
	if job != nil {
		t.Errorf("Expected no more queued jobs, but found job %s", job.ID)
	}
}

func TestQueueService_ProcessMixedStatusJobs(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)

	// Create a test account
	account, err := dbService.CreateAccount("test-account", "folder-123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	// Create jobs with different statuses
	queuedJobs := make([]string, 0)

	// Create completed job
	completedJob, _ := dbService.CreateJob(account.ID, "completed.epub")
	_ = dbService.MarkJobCompleted(completedJob.ID, "processed-completed.epub", "http://drive.example.com/completed")

	// Create queued job 1
	job1, _ := dbService.CreateJob(account.ID, "queued1.epub")
	queuedJobs = append(queuedJobs, job1.ID)
	time.Sleep(10 * time.Millisecond)

	// Create failed job
	failedJob, _ := dbService.CreateJob(account.ID, "failed.epub")
	_ = dbService.MarkJobFailed(failedJob.ID, "test error")

	// Create queued job 2
	job2, _ := dbService.CreateJob(account.ID, "queued2.epub")
	queuedJobs = append(queuedJobs, job2.ID)
	time.Sleep(10 * time.Millisecond)

	// Create processing job (simulating a stuck job)
	processingJob, _ := dbService.CreateJob(account.ID, "processing.epub")
	processingJob.Status = "processing"
	_ = dbService.UpdateJob(processingJob)

	// Create queued job 3
	job3, _ := dbService.CreateJob(account.ID, "queued3.epub")
	queuedJobs = append(queuedJobs, job3.ID)

	// Track processed jobs
	processedOrder := make([]string, 0)

	// Process jobs using GetNextQueuedJob - should only get queued ones
	for i := 0; i < 10; i++ { // Run more times than needed
		job, err := dbService.GetNextQueuedJob()
		if err != nil {
			t.Fatalf("GetNextQueuedJob() error = %v", err)
		}
		if job == nil {
			break // No more queued jobs
		}

		processedOrder = append(processedOrder, job.ID)

		// Simulate processing by marking as completed
		err = dbService.MarkJobCompleted(job.ID, "processed-"+job.OriginalFilename, "http://drive.example.com/file")
		if err != nil {
			t.Errorf("Failed to mark job completed: %v", err)
		}
	}

	// Verify only queued jobs were processed
	if len(processedOrder) != 3 {
		t.Errorf("Expected 3 queued jobs to be processed, got %d", len(processedOrder))
	}

	// Verify correct jobs were processed in FIFO order
	for i, jobID := range processedOrder {
		if jobID != queuedJobs[i] {
			t.Errorf("Expected job %d to be %s, got %s", i, queuedJobs[i], jobID)
		}
	}
}

func TestQueueService_ProcessJobsSequentially(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)

	// Create a test account
	account, err := dbService.CreateAccount("test-account", "folder-123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	// Create a job
	job, err := dbService.CreateJob(account.ID, "test.epub")
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Verify initial status
	fetchedJob, _ := dbService.GetJob(job.ID)
	if fetchedJob.Status != "queued" {
		t.Errorf("Expected initial status to be 'queued', got %s", fetchedJob.Status)
	}

	// Track status transitions by simulating the processing workflow
	statusTransitions := make([]string, 0)

	// Get the queued job
	queuedJob, err := dbService.GetNextQueuedJob()
	if err != nil {
		t.Fatalf("GetNextQueuedJob() error = %v", err)
	}
	if queuedJob == nil || queuedJob.ID != job.ID {
		t.Fatalf("Expected to get job %s", job.ID)
	}

	// Simulate processing - update to processing status
	queuedJob.Status = "processing"
	err = dbService.UpdateJob(queuedJob)
	if err != nil {
		t.Fatalf("Failed to update job to processing: %v", err)
	}

	fetchedJob, _ = dbService.GetJob(job.ID)
	statusTransitions = append(statusTransitions, fetchedJob.Status)

	// Simulate work
	time.Sleep(10 * time.Millisecond)

	// Mark as completed
	err = dbService.MarkJobCompleted(job.ID, "processed-"+job.OriginalFilename, "http://drive.example.com/file")
	if err != nil {
		t.Fatalf("Failed to mark job completed: %v", err)
	}

	fetchedJob, _ = dbService.GetJob(job.ID)
	statusTransitions = append(statusTransitions, fetchedJob.Status)

	// Verify status transitions
	expectedTransitions := []string{"processing", "completed"}
	if len(statusTransitions) != len(expectedTransitions) {
		t.Errorf("Expected %d status transitions, got %d", len(expectedTransitions), len(statusTransitions))
	}

	for i, status := range statusTransitions {
		if i < len(expectedTransitions) && status != expectedTransitions[i] {
			t.Errorf("Expected transition %d to be '%s', got '%s'", i, expectedTransitions[i], status)
		}
	}

	// Verify final state
	finalJob, _ := dbService.GetJob(job.ID)
	if finalJob.Status != "completed" {
		t.Errorf("Expected final status to be 'completed', got %s", finalJob.Status)
	}
	if finalJob.ProcessedFilename == "" {
		t.Errorf("Expected ProcessedFilename to be set")
	}
	if finalJob.DriveURL == "" {
		t.Errorf("Expected DriveURL to be set")
	}
}
