package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"Backup_image/analyzer"
	"Backup_image/model"
	"Backup_image/repository"
	"github.com/joho/godotenv"
)

type AnalyzedImage struct {
	File  string   `json:"file"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	Error string   `json:"error,omitempty"`
}

type ImageService interface {
	SetContext(context.Context)
	AnalyzeFolder(folder string, progress func(done, total int, file string)) ([]AnalyzedImage, error)
	GetTagCloud() ([]model.TagCount, error)
	SearchImages(keyword, tag string) ([]model.Image, error)
	DeleteImageRecord(id uint) error
}

type imageService struct {
	ctx  context.Context
	repo repository.ImageRepository
}

func NewImageService(repo repository.ImageRepository) ImageService {
	return &imageService{repo: repo}
}

func (s *imageService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *imageService) context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

var (
	loadEnvOnce sync.Once
	loadEnvErr  error
)

func newAnalyzer() (analyzer.Analyzer, error) {
	loadEnvOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
			loadEnvErr = err
		}
	})
	if loadEnvErr != nil {
		return nil, fmt.Errorf("load .env: %w", loadEnvErr)
	}
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ยังไม่ได้ตั้ง GEMINI_API_KEY (ใส่ในไฟล์ .env หรือใช้ setx)")
	}
	return analyzer.NewGemini(key, os.Getenv("GEMINI_MODEL")), nil
}

func (s *imageService) AnalyzeFolder(folder string, progress func(done, total int, file string)) ([]AnalyzedImage, error) {
	an, err := newAnalyzer()
	if err != nil {
		return nil, err
	}
	folder, err = filepath.Abs(folder)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบโฟลเดอร์ไม่ได้: %w", err)
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("อ่านโฟลเดอร์ไม่ได้: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && analyzer.IsSupported(entry.Name()) {
			files = append(files, filepath.Join(folder, entry.Name()))
		}
	}
	results := make([]AnalyzedImage, 0, len(files))
	for i, path := range files {
		ctx, cancel := context.WithTimeout(s.context(), 3*time.Minute)
		result, analyzeErr := an.Analyze(ctx, path)
		cancel()

		item := AnalyzedImage{File: filepath.Base(path), Path: path, Tags: result.Tags}
		if analyzeErr != nil {
			item.Error = analyzeErr.Error()
			item.Tags = []string{}
		} else {
			item.Tags = analyzer.NormalizeTags(result.Tags, 10)
			dbCtx, dbCancel := databaseContext(s.context())
			_, saveErr := s.repo.UpsertAnalyzedImage(dbCtx, path, item.Tags)
			dbCancel()
			if saveErr != nil {
				item.Error = fmt.Sprintf("วิเคราะห์ได้แต่บันทึกแท็กไม่สำเร็จ: %v", saveErr)
				item.Tags = []string{}
			}
		}
		results = append(results, item)
		if progress != nil {
			progress(i+1, len(files), item.File)
		}
	}
	return results, nil
}

func (s *imageService) GetTagCloud() ([]model.TagCount, error) {
	ctx, cancel := databaseContext(s.context())
	defer cancel()
	return s.repo.GetTagCloud(ctx)
}

func (s *imageService) SearchImages(keyword, tag string) ([]model.Image, error) {
	ctx, cancel := databaseContext(s.context())
	defer cancel()
	return s.repo.SearchImages(ctx, keyword, tag)
}

func (s *imageService) DeleteImageRecord(id uint) error {
	ctx, cancel := databaseContext(s.context())
	defer cancel()
	return s.repo.DeleteImageRecord(ctx, id)
}
