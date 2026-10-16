package main

import (
	"Backup_image/model"
	"Backup_image/repository"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx   context.Context
	db    *gorm.DB
	dbErr error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := repository.NewDbConnection()
	if err != nil {
		a.dbErr = err
		return
	}

	a.db = db
}
func moveOneFile(sourcePath string, destPath string) error {
	// ถ้ามีไฟล์ชื่อเดียวกันที่ปลายทาง ห้ามเขียนทับ
	//check ซ้ำ
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("ไฟล์ปลายทางมีอยู่แล้ว: %s", destPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("ตรวจไฟล์ปลายทางไม่ได้: %w", err)
	}

	// ลองย้ายก่อน ถ้าข้ามไดรฟ์ไม่ได้จึงค่อยคัดลอก
	if err := os.Rename(sourcePath, destPath); err == nil {
		return nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("เปิดไฟล์ต้นทางไม่ได้: %w", err)
	}

	destFile, err := os.OpenFile(
		destPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		sourceFile.Close()
		return fmt.Errorf("สร้างไฟล์ปลายทางไม่ได้: %w", err)
	}

	_, copyErr := io.Copy(destFile, sourceFile)
	sourceCloseErr := sourceFile.Close()
	destCloseErr := destFile.Close()

	if copyErr != nil {
		os.Remove(destPath)
		return fmt.Errorf("คัดลอกไฟล์ไม่สำเร็จ: %w", copyErr)
	}
	if sourceCloseErr != nil {
		os.Remove(destPath)
		return fmt.Errorf("ปิดไฟล์ต้นทางไม่สำเร็จ: %w", sourceCloseErr)
	}
	if destCloseErr != nil {
		os.Remove(destPath)
		return fmt.Errorf("ปิดไฟล์ปลายทางไม่สำเร็จ: %w", destCloseErr)
	}

	// ลบต้นทางหลังจากคัดลอกและปิดไฟล์สำเร็จแล้วเท่านั้น
	if err := os.Remove(sourcePath); err != nil {
		os.Remove(destPath)
		return fmt.Errorf("ลบไฟล์ต้นทางไม่สำเร็จ: %w", err)
	}

	return nil
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
func (a *App) MoveFiles(sourcePath string, destPath string) (int, error) {
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

	destInfo, err := os.Stat(destAbsolute)
	if err != nil {
		return 0, fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}
	if !destInfo.IsDir() {
		return 0, fmt.Errorf("ปลายทางไม่ใช่โฟลเดอร์")
	}

	entries, err := os.ReadDir(sourceAbsolute)
	if err != nil {
		return 0, fmt.Errorf("อ่านโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}

	moved := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		sourceFile := filepath.Join(sourceAbsolute, entry.Name())
		destFile := filepath.Join(destAbsolute, entry.Name())

		if _, err := os.Stat(destFile); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return moved, fmt.Errorf("ตรวจไฟล์ปลายทางไม่ได้: %w", err)
		}
		var oldRecord model.BackupRecord
		err := a.db.
			Where("destination_path = ? AND file_name = ?", destAbsolute, entry.Name()).
			First(&oldRecord).Error

		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return moved, fmt.Errorf("ตรวจประวัติไฟล์ไม่ได้: %w", err)
		}

		if err := moveOneFile(sourceFile, destFile); err != nil {
			return moved, fmt.Errorf("ย้ายไฟล์ %s ไม่สำเร็จ: %w", entry.Name(), err)
		}

		record := model.BackupRecord{
			SourcePath:      sourceFile,
			DestinationPath: destAbsolute,
			FileName:        entry.Name(),
			Status:          "moved",
		}

		if err := a.db.Create(&record).Error; err != nil {
			return moved, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %w", entry.Name(), err)
		}
		moved++
	}
	return moved, nil
}

type MoveResult struct {
	Moved      int   `json:"moved"`
	DurationMs int64 `json:"durationMs"`
}

func (a *App) MoveFilesTimed(sourcePath string, destPath string) (MoveResult, error) {
	start := time.Now()

	moved, err := a.MoveFiles(sourcePath, destPath)

	result := MoveResult{
		Moved:      moved,
		DurationMs: time.Since(start).Milliseconds(),
	}

	return result, err
}
func (a *App) GetHistory(destinationPath string) ([]model.BackupRecord, error) {
	destinationPath = strings.TrimSpace(destinationPath)
	if destinationPath == "" {
		return nil, fmt.Errorf("กรุณาระบุโฟลเดอร์ปลายทาง")
	}

	destinationAbsolute, err := filepath.Abs(destinationPath)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}

	if a.dbErr != nil {
		return nil, fmt.Errorf("เปิดฐานข้อมูลไม่ได้: %w", a.dbErr)
	}
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	var records []model.BackupRecord
	err = a.db.
		Where("destination_path = ?", destinationAbsolute).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("อ่านประวัติไม่ได้: %w", err)
	}

	return records, nil
}
func (a *App) FindMissingFiles(destinationPath string) ([]string, error) {
	destinationPath = strings.TrimSpace(destinationPath)
	if destinationPath == "" {
		return nil, fmt.Errorf("กรุณาระบุโฟลเดอร์ปลายทาง")
	}

	destinationAbsolute, err := filepath.Abs(destinationPath)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}

	if a.dbErr != nil {
		return nil, fmt.Errorf("เปิดฐานข้อมูลไม่ได้: %w", a.dbErr)
	}
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	var records []model.BackupRecord
	err = a.db.
		Where("destination_path = ? AND status = ?", destinationAbsolute, "moved").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("อ่านประวัติไม่ได้: %w", err)
	}

	missingFiles := []string{}
	for _, record := range records {
		filePath := filepath.Join(destinationAbsolute, record.FileName)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, record.FileName)
		} else if err != nil {
			return nil, fmt.Errorf("ตรวจไฟล์ %s ไม่ได้: %w", record.FileName, err)
		}
	}

	return missingFiles, nil
}
