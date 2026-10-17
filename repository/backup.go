package repository

import (
	"context"
	"errors"

	"Backup_image/model"
	"gorm.io/gorm"
)

var ErrDatabaseUnavailable = errors.New("database is not available")

type BackupRepository interface {
	Create(ctx context.Context, record *model.BackupRecord) error
	FindByDestinationAndFile(ctx context.Context, destination, fileName string) (model.BackupRecord, bool, error)
	ListByDestination(ctx context.Context, destination string) ([]model.BackupRecord, error)
	ListMissingCandidates(ctx context.Context, destination string) ([]model.BackupRecord, error)
}

type backupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) BackupRepository {
	return &backupRepository{db: db}
}

func (r *backupRepository) Create(ctx context.Context, record *model.BackupRecord) error {
	if r.db == nil {
		return ErrDatabaseUnavailable
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *backupRepository) FindByDestinationAndFile(ctx context.Context, destination, fileName string) (model.BackupRecord, bool, error) {
	if r.db == nil {
		return model.BackupRecord{}, false, ErrDatabaseUnavailable
	}
	var record model.BackupRecord
	err := r.db.WithContext(ctx).
		Where("destination_path = ? AND file_name = ?", destination, fileName).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return record, false, nil
	}
	return record, err == nil, err
}

func (r *backupRepository) ListByDestination(ctx context.Context, destination string) ([]model.BackupRecord, error) {
	if r.db == nil {
		return nil, ErrDatabaseUnavailable
	}
	var records []model.BackupRecord
	err := r.db.WithContext(ctx).
		Where("destination_path = ?", destination).
		Order("created_at DESC, id DESC").
		Find(&records).Error
	return records, err
}

func (r *backupRepository) ListMissingCandidates(ctx context.Context, destination string) ([]model.BackupRecord, error) {
	if r.db == nil {
		return nil, ErrDatabaseUnavailable
	}
	var records []model.BackupRecord
	err := r.db.WithContext(ctx).
		Where("destination_path = ? AND status IN ?", destination, []string{"moved", "copied"}).
		Find(&records).Error
	return records, err
}
