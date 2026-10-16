package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
func (a *App) SetFolder(target string, input string) (string, error) {
	if target != "source" && target != "dest" {
		return "", fmt.Errorf("target ต้องเป็น source หรือ dest")
	}

	path := strings.TrimSpace(input)

	// ตัดอัญประกาศที่ติดมากับ path เวลาคัดลอกมาวาง
	if len(path) >= 2 &&
		((path[0] == '"' && path[len(path)-1] == '"') ||
			(path[0] == '\'' && path[len(path)-1] == '\'')) {
		path = strings.TrimSpace(path[1 : len(path)-1])
	}

	if path == "" {
		return "", fmt.Errorf("กรุณาระบุ path ของโฟลเดอร์")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("แปลง path ไม่สำเร็จ: %w", err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("ตรวจสอบโฟลเดอร์ไม่ได้: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path นี้ไม่ใช่โฟลเดอร์: %s", absolutePath)
	}

	directory, err := os.Open(absolutePath)
	if err != nil {
		return "", fmt.Errorf("เปิดโฟลเดอร์ไม่ได้: %w", err)
	}
	defer directory.Close()

	return absolutePath, nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
func (a *App) ListFiles(path string) ([]string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่าโฟลเดอร์ก่อน")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("อ่านโฟลเดอร์ไม่ได้: %w", err)
	}

	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name()) //ถ้าไม่ใช่ folder ให้เอาชื่ออกมา
		}
	}

	return files, nil
}

// คัดลอกไฟล์
func (a *App) CopyFiles(sourcePath string, destPath string) (int, error) {
	sourcePath = strings.TrimSpace(sourcePath)
	destPath = strings.TrimSpace(destPath)

	if sourcePath == "" || destPath == "" {
		return 0, fmt.Errorf("กรุณาตั้งค่าโฟลเดอร์ต้นทางและปลายทางก่อน")
	}

	sourceAbsolute, err := filepath.Abs(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("ตรวจสอบโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}

	destAbsolute, err := filepath.Abs(destPath)
	if err != nil {
		return 0, fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}

	if sourceAbsolute == destAbsolute {
		return 0, fmt.Errorf("โฟลเดอร์ต้นทางกับปลายทางต้องไม่ใช่โฟลเดอร์เดียวกัน")
	}

	entries, err := os.ReadDir(sourceAbsolute)
	if err != nil {
		return 0, fmt.Errorf("อ่านโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}

	copied := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		//เอา path มาต่อกับชื่อไฟล์จะได้ paht เต็ม
		sourceFilePath := filepath.Join(sourceAbsolute, entry.Name())
		destFilePath := filepath.Join(destAbsolute, entry.Name())

		sourceFile, err := os.Open(sourceFilePath)
		if err != nil {
			return copied, fmt.Errorf("เปิดไฟล์ %s ไม่ได้: %w", entry.Name(), err)
		}

		destFile, err := os.Create(destFilePath)
		if err != nil {
			sourceFile.Close()
			return copied, fmt.Errorf("สร้างไฟล์ปลายทาง %s ไม่ได้: %w", entry.Name(), err)
		}
		//คัดลอกข้อมูลจากไฟล์ต้นทางไปไฟล์ปลายทาง แล้วปิดไฟล์ทั้งสอง
		_, copyErr := io.Copy(destFile, sourceFile)
		sourceCloseErr := sourceFile.Close()
		destCloseErr := destFile.Close()

		if copyErr != nil {
			return copied, fmt.Errorf("คัดลอกไฟล์ %s ไม่สำเร็จ: %w", entry.Name(), copyErr)
		}
		if sourceCloseErr != nil {
			return copied, fmt.Errorf("ปิดไฟล์ต้นทาง %s ไม่สำเร็จ: %w", entry.Name(), sourceCloseErr)
		}
		if destCloseErr != nil {
			return copied, fmt.Errorf("ปิดไฟล์ปลายทาง %s ไม่สำเร็จ: %w", entry.Name(), destCloseErr)
		}

		copied++
	}

	return copied, nil
}
