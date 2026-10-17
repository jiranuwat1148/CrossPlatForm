package main

import (
	"context"
	"fmt"
	"time"

	"Backup_image/model"
	"Backup_image/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	backupService service.BackupService
	imageService  service.ImageService
	dbErr         error
}

type MoveResult struct {
	Moved      int   `json:"moved"`
	DurationMs int64 `json:"durationMs"`
}

func NewApp(backupService service.BackupService, imageService service.ImageService, dbErr error) *App {
	return &App{backupService: backupService, imageService: imageService, dbErr: dbErr}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.backupService != nil {
		a.backupService.SetContext(ctx)
	}
	if a.imageService != nil {
		a.imageService.SetContext(ctx)
	}
}

func (a *App) requireDatabase() error {
	if a.dbErr != nil {
		return fmt.Errorf("เปิดฐานข้อมูลไม่ได้: %w", a.dbErr)
	}
	return nil
}

func (a *App) requireBackupService() error {
	if a.backupService == nil {
		return fmt.Errorf("บริการจัดการไฟล์ยังไม่พร้อม")
	}
	return nil
}

func (a *App) requireBackupDatabase() error {
	if err := a.requireDatabase(); err != nil {
		return err
	}
	return a.requireBackupService()
}

func (a *App) requireImageService() error {
	if a.imageService == nil {
		return fmt.Errorf("บริการวิเคราะห์รูปยังไม่พร้อม")
	}
	return nil
}

func (a *App) requireImageDatabase() error {
	if err := a.requireDatabase(); err != nil {
		return err
	}
	return a.requireImageService()
}

func (a *App) SetFolder(target, input string) (string, error) {
	if err := a.requireBackupService(); err != nil {
		return "", err
	}
	return a.backupService.SetFolder(target, input)
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) ListFiles(path string) ([]string, error) {
	if err := a.requireBackupService(); err != nil {
		return nil, err
	}
	return a.backupService.ListFiles(path)
}

func (a *App) CopyFiles(sourcePath, destPath string) (int, error) {
	if err := a.requireBackupDatabase(); err != nil {
		return 0, err
	}
	return a.backupService.CopyFiles(sourcePath, destPath)
}

func (a *App) MoveFiles(sourcePath, destPath string) (int, error) {
	if err := a.requireBackupDatabase(); err != nil {
		return 0, err
	}
	return a.backupService.MoveFiles(sourcePath, destPath)
}

func (a *App) MoveFilesTimed(sourcePath, destPath string) (MoveResult, error) {
	start := time.Now()
	moved, err := a.MoveFiles(sourcePath, destPath)
	return MoveResult{Moved: moved, DurationMs: time.Since(start).Milliseconds()}, err
}

func (a *App) GetHistory(destinationPath string) ([]model.BackupRecord, error) {
	if err := a.requireBackupDatabase(); err != nil {
		return nil, err
	}
	return a.backupService.GetHistory(destinationPath)
}

func (a *App) FindMissingFiles(destinationPath string) ([]string, error) {
	if err := a.requireBackupDatabase(); err != nil {
		return nil, err
	}
	return a.backupService.FindMissingFiles(destinationPath)
}

func (a *App) GetTagCloud() ([]model.TagCount, error) {
	if err := a.requireImageDatabase(); err != nil {
		return nil, err
	}
	return a.imageService.GetTagCloud()
}

func (a *App) AnalyzeFolder(folder string) ([]ImageTags, error) {
	if err := a.requireImageDatabase(); err != nil {
		return nil, err
	}

	results, err := a.imageService.AnalyzeFolder(folder, func(done, total int, file string) {
		runtime.EventsEmit(a.ctx, "analyze:progress", AnalyzeProgress{Done: done, Total: total, File: file})
	})
	if err != nil {
		return nil, err
	}

	images := make([]ImageTags, 0, len(results))
	for _, result := range results {
		images = append(images, ImageTags{File: result.File, Path: result.Path, Tags: result.Tags, Error: result.Error})
	}
	return images, nil
}
