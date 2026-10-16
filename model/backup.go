package model

import "time"

type BackupRecord struct {
	ID              uint   `gorm:"primaryKey"`
	SourcePath      string `gorm:"not null"`
	DestinationPath string `gorm:"not null;index"`
	FileName        string `gorm:"not null;index"`
	Status          string `gorm:"not null;default:moved"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
