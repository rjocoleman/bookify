package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bookify/internal/db"
)

type QueueService struct {
	db        *db.Service
	drive     *DriveService
	processor *ProcessorService
	tempDir   string
	stopCh    chan bool
}

func NewQueueService(dbService *db.Service, driveService *DriveService) *QueueService {
	return &QueueService{
		db:        dbService,
		drive:     driveService,
		processor: NewProcessorService(),
		tempDir:   "./temp",
		stopCh:    make(chan bool),
	}
}

func (q *QueueService) StartWorker() {
	log.Println("Starting queue worker...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			q.processNextJob()
		case <-q.stopCh:
			log.Println("Queue worker stopped")
			return
		}
	}
}

func (q *QueueService) Stop() {
	q.stopCh <- true
}

func (q *QueueService) processNextJob() {
	job, err := q.db.GetNextQueuedJob()
	if err != nil {
		log.Printf("Failed to get next queued job: %v", err)
		return
	}

	if job == nil {
		return
	}

	log.Printf("Processing job %s: %s", job.ID, job.OriginalFilename)
	q.processJob(job)
}

func (q *QueueService) processJob(job *db.Job) {
	job.Status = "processing"
	job.Stage = "starting"
	job.Progress = 0
	if err := q.db.UpdateJob(job); err != nil {
		log.Printf("Warning: Failed to update job: %v", err)
	}

	inputPath := filepath.Join(q.tempDir, job.OriginalFilename)
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		q.failJob(job, "Input file not found")
		return
	}

	job.Stage = "converting"
	job.Progress = 25
	if err := q.db.UpdateJob(job); err != nil {
		log.Printf("Warning: Failed to update job: %v", err)
	}

	outputPath, err := q.processor.PrepareOutputPath(q.tempDir, job.OriginalFilename)
	if err != nil {
		q.failJob(job, fmt.Sprintf("Failed to prepare output path: %v", err))
		return
	}

	err = q.processor.ProcessEPUB(inputPath, outputPath, func(progress int) {
		job.Progress = 25 + (progress * 50 / 100)
		if err := q.db.UpdateJob(job); err != nil {
			log.Printf("Warning: Failed to update job: %v", err)
		}
	})

	if err != nil {
		q.failJob(job, fmt.Sprintf("Conversion failed: %v", err))
		return
	}

	job.Stage = "uploading"
	job.Progress = 75
	if err := q.db.UpdateJob(job); err != nil {
		log.Printf("Warning: Failed to update job: %v", err)
	}

	account, err := q.db.GetAccount(job.AccountID)
	if err != nil {
		q.failJob(job, fmt.Sprintf("Failed to get account: %v", err))
		return
	}

	cleanFilename := q.processor.CleanFilename(job.OriginalFilename)
	driveURL, err := q.drive.UploadFile(account, outputPath, cleanFilename)
	if err != nil {
		q.failJob(job, fmt.Sprintf("Upload failed: %v", err))
		return
	}

	job.Stage = "cleanup"
	job.Progress = 90
	if err := q.db.UpdateJob(job); err != nil {
		log.Printf("Warning: Failed to update job: %v", err)
	}

	if err := os.Remove(inputPath); err != nil {
		log.Printf("Warning: Failed to remove input file: %v", err)
	}
	if err := os.Remove(outputPath); err != nil {
		log.Printf("Warning: Failed to remove output file: %v", err)
	}

	err = q.db.MarkJobCompleted(job.ID, cleanFilename, driveURL)
	if err != nil {
		log.Printf("Failed to mark job completed: %v", err)
	}

	log.Printf("Job %s completed successfully", job.ID)
}

func (q *QueueService) failJob(job *db.Job, errorMsg string) {
	log.Printf("Job %s failed: %s", job.ID, errorMsg)
	if err := q.db.MarkJobFailed(job.ID, errorMsg); err != nil {
		log.Printf("Warning: Failed to mark job as failed: %v", err)
	}
}

func (q *QueueService) StartCleanupWorker() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				q.cleanupOldFiles()
			case <-q.stopCh:
				return
			}
		}
	}()
}

func (q *QueueService) cleanupOldFiles() {
	cutoff := time.Now().Add(-1 * time.Hour)

	entries, err := os.ReadDir(q.tempDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(q.tempDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: Failed to remove old file %s: %v", filePath, err)
			}
			log.Printf("Cleaned up old temp file: %s", entry.Name())
		}
	}
}
