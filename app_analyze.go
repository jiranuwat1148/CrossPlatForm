package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"Backup_image/analyzer"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ImageTags struct {
	File  string   `json:"file"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	Error string   `json:"error,omitempty"`
}

type AnalyzeProgress struct {
	Done  int    `json:"done"`
	Total int    `json:"total"`
	File  string `json:"file"`
}

var loadEnvOnce sync.Once

func newAnalyzer() (analyzer.Analyzer, error) {
	loadEnvOnce.Do(func() { _ = godotenv.Load() })

	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ยังไม่ได้ตั้ง GEMINI_API_KEY (ใส่ในไฟล์ .env หรือใช้ setx)")
	}
	return analyzer.NewGemini(key, os.Getenv("GEMINI_MODEL")), nil
}
func (a *App) AnalyzeFolder(folder string) ([]ImageTags, error) {
	an, err := newAnalyzer()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("อ่านโฟลเดอร์ไม่ได้: %w", err)
	}

	// เลือกเฉพาะไฟล์รูปที่รองรับ
	var files []string
	for _, e := range entries {
		if !e.IsDir() && analyzer.IsSupported(e.Name()) {
			files = append(files, filepath.Join(folder, e.Name()))
		}
	}

	results := make([]ImageTags, 0, len(files))

	for i, path := range files {
		ctx, cancel := context.WithTimeout(a.ctx, 3*time.Minute)
		r, err := an.Analyze(ctx, path)
		cancel()

		item := ImageTags{
			File: filepath.Base(path),
			Path: path,
			Tags: r.Tags,
		}
		if err != nil {
			item.Error = err.Error()
			item.Tags = []string{}
		}
		results = append(results, item)

		runtime.EventsEmit(a.ctx, "analyze:progress", AnalyzeProgress{
			Done:  i + 1,
			Total: len(files),
			File:  item.File,
		})
	}

	return results, nil
}
