package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListAccounts() ([]Account, error) {
	var accounts []Account
	err := s.db.Find(&accounts).Error
	return accounts, err
}

func (s *Service) CreateAccount(name, folderID string) (*Account, error) {
	account := &Account{
		Name:     name,
		FolderID: folderID,
	}
	err := s.db.Create(account).Error
	return account, err
}

func (s *Service) CreateAccountWithOAuth(account *Account) error {
	return s.db.Create(account).Error
}

func (s *Service) UpdateAccount(account *Account) error {
	return s.db.Save(account).Error
}

func (s *Service) GetAccount(id uint) (*Account, error) {
	var account Account
	err := s.db.First(&account, id).Error
	return &account, err
}

func (s *Service) GetAccountByName(name string) (*Account, error) {
	var account Account
	err := s.db.Where("name = ?", name).First(&account).Error
	return &account, err
}

func (s *Service) CreateJob(accountID uint, originalFilename string) (*Job, error) {
	job := &Job{
		ID:               uuid.New().String(),
		AccountID:        accountID,
		OriginalFilename: originalFilename,
		Status:           "queued",
		Stage:            "queued",
		Progress:         0,
	}
	err := s.db.Create(job).Error
	return job, err
}

func (s *Service) UpdateJob(job *Job) error {
	return s.db.Save(job).Error
}

func (s *Service) GetJob(id string) (*Job, error) {
	var job Job
	err := s.db.Preload("Account").First(&job, "id = ?", id).Error
	return &job, err
}

func (s *Service) ListRecentJobs(limit int) ([]Job, error) {
	var jobs []Job
	err := s.db.Preload("Account").Order("created_at desc").Limit(limit).Find(&jobs).Error
	return jobs, err
}

func (s *Service) MarkJobCompleted(jobID string, processedFilename, driveURL string) error {
	now := time.Now()
	return s.db.Model(&Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":             "completed",
		"stage":              "completed",
		"progress":           100,
		"processed_filename": processedFilename,
		"drive_url":          driveURL,
		"completed_at":       &now,
	}).Error
}

func (s *Service) MarkJobFailed(jobID string, errorMsg string) error {
	return s.db.Model(&Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status": "failed",
		"stage":  "failed",
		"error":  errorMsg,
	}).Error
}
