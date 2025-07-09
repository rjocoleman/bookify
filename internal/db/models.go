package db

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Account struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;not null" json:"name"`
	FolderID     string    `gorm:"not null" json:"folder_id"`
	AccessToken  string    `gorm:"size:2048" json:"-"`
	RefreshToken string    `gorm:"size:512" json:"-"`
	TokenExpiry  time.Time `json:"-"`
	UserEmail    string    `json:"user_email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Jobs         []Job     `gorm:"foreignKey:AccountID" json:"-"`
}

type Job struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	AccountID         uint       `gorm:"not null" json:"account_id"`
	Account           Account    `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	OriginalFilename  string     `gorm:"not null" json:"original_filename"`
	ProcessedFilename string     `json:"processed_filename"`
	Status            string     `gorm:"not null;default:queued" json:"status"`
	Progress          int        `gorm:"default:0" json:"progress"`
	Stage             string     `gorm:"default:queued" json:"stage"`
	Message           string     `json:"message"`
	DriveURL          string     `json:"drive_url"`
	Error             string     `json:"error"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CompletedAt       *time.Time `json:"completed_at"`
}

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Account{}, &Job{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
