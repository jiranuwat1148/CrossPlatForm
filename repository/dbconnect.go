package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Backup_image/model"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func NewDbConnection() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	dbPath := strings.TrimSpace(os.Getenv("DATABASE"))
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config directory: %w", err)
	}
	appDataDir := filepath.Join(configDir, "Backup_image")
	if dbPath == "" {
		dbPath = filepath.Join(appDataDir, "image.db")
	} else if !filepath.IsAbs(dbPath) {
		dbPath, err = filepath.Abs(dbPath)
		if err != nil {
			return nil, fmt.Errorf("resolve database path: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := database.AutoMigrate(
		&model.Image{},
		&model.Tag{},
		&model.BackupRecord{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := database.Exec("CREATE INDEX IF NOT EXISTS idx_image_tags_tag_id ON image_tags(tag_id)").Error; err != nil {
		return nil, fmt.Errorf("index image tags: %w", err)
	}

	return database, nil
}
