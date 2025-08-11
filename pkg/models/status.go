package models

import "time"

type Status struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	DatabaseName   string `json:"database_name" gorm:"unique"`
	LastBackupDate string `json:"last_backup_date"`
	BackupSize     int64  `json:"backup_size"`
	BackupCount    int    `json:"backup_count"`
	S3Link         string `json:"s3_link"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `sql:"index"`
	LastRunStatus  string     `json:"last_run_status"`
}
