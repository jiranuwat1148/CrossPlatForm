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
	if dbPath == "" {
		defaultPath, err := defaultDatabasePath()
		if err != nil {
			return nil, fmt.Errorf("resolve default database path: %w", err)
		}
		dbPath = defaultPath
	}
	if !filepath.IsAbs(dbPath) {
		absolutePath, err := filepath.Abs(dbPath)
		if err != nil {
			return nil, fmt.Errorf("resolve database path: %w", err)
		}
		dbPath = absolutePath
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

func defaultDatabasePath() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	if path := findDatabaseFrom(workingDirectory); path != "" {
		return path, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}
	if path := findDatabaseFrom(filepath.Dir(executable)); path != "" {
		return path, nil
	}

	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config directory: %w", err)
	}
	return filepath.Join(configDirectory, "Backup_image", "image.db"), nil
}

func findDatabaseFrom(directory string) string {
	for range 5 {
		candidate := filepath.Join(directory, "database", "image.db")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return ""
}
