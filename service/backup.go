package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"Backup_image/model"
	"Backup_image/repository"
)

type BackupService interface {
	SetContext(context.Context)
	SetFolder(target, input string) (string, error)
	ListFiles(path string) ([]string, error)
	CopyFiles(sourcePath, destPath string) (int, error)
	MoveFiles(sourcePath, destPath string) (int, error)
	CopySelectedFiles(sourcePath, destPath string, fileNames []string) (int, error)
	MoveSelectedFiles(sourcePath, destPath string, fileNames []string) (int, error)
	DeleteSelectedFiles(destPath string, fileNames []string) (int, error)
	GetHistory(destinationPath string) ([]model.BackupRecord, error)
	FindMissingFiles(destinationPath string) ([]string, error)
}

type backupService struct {
	ctx  context.Context
	repo repository.BackupRepository
}

func NewBackupService(repo repository.BackupRepository) BackupService {
	return &backupService{repo: repo}
}

func (s *backupService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *backupService) context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

func (s *backupService) SetFolder(target, input string) (string, error) {
	if target != "source" && target != "dest" {
		return "", fmt.Errorf("target ต้องเป็น source หรือ dest")
	}

	path := strings.TrimSpace(input)
	if len(path) >= 2 && ((path[0] == '"' && path[len(path)-1] == '"') || (path[0] == '\'' && path[len(path)-1] == '\'')) {
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

func (s *backupService) ListFiles(path string) ([]string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่าโฟลเดอร์ก่อน")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("อ่านโฟลเดอร์ไม่ได้: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func (s *backupService) CopyFiles(sourcePath, destPath string) (int, error) {
	sourceAbsolute, _, err := validateFolders(sourcePath, destPath)
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(sourceAbsolute)
	if err != nil {
		return 0, fmt.Errorf("อ่านโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return 0, fmt.Errorf("ตรวจไฟล์ %s ไม่ได้: %w", entry.Name(), err)
		}
		if info.Mode().IsRegular() {
			fileNames = append(fileNames, entry.Name())
		}
	}
	if len(fileNames) == 0 {
		return 0, nil
	}
	return s.CopySelectedFiles(sourcePath, destPath, fileNames)
}

func (s *backupService) MoveFiles(sourcePath, destPath string) (int, error) {
	sourceAbsolute, destAbsolute, err := validateFolders(sourcePath, destPath)
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(sourceAbsolute)
	if err != nil {
		return 0, fmt.Errorf("อ่านโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}

	moved := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return moved, fmt.Errorf("ตรวจไฟล์ %s ไม่ได้: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		destFile := filepath.Join(destAbsolute, entry.Name())
		if _, err := os.Lstat(destFile); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return moved, fmt.Errorf("ตรวจไฟล์ปลายทางไม่ได้: %w", err)
		}
		ctx, cancel := databaseContext(s.context())
		_, found, err := s.repo.FindByDestinationAndFile(ctx, destAbsolute, entry.Name())
		cancel()
		if err != nil {
			return moved, fmt.Errorf("ตรวจประวัติไฟล์ไม่ได้: %w", err)
		}
		if found {
			continue
		}
		if err := s.moveOneFile(filepath.Join(sourceAbsolute, entry.Name()), destFile); err != nil {
			return moved, fmt.Errorf("ย้ายไฟล์ %s ไม่สำเร็จ: %w", entry.Name(), err)
		}
		record := model.BackupRecord{SourcePath: filepath.Join(sourceAbsolute, entry.Name()), DestinationPath: destAbsolute, FileName: entry.Name(), Status: "moved"}
		ctx, cancel = databaseContext(s.context())
		err = s.repo.Create(ctx, &record)
		cancel()
		if err != nil {
			if rollbackErr := s.moveOneFile(destFile, record.SourcePath); rollbackErr != nil {
				return moved, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %v; และย้ายไฟล์กลับต้นทางไม่ได้: %w", entry.Name(), err, rollbackErr)
			}
			return moved, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %w", entry.Name(), err)
		}
		moved++
	}
	return moved, nil
}

func (s *backupService) CopySelectedFiles(sourcePath, destPath string, fileNames []string) (int, error) {
	sourceAbsolute, destAbsolute, err := validateSelectedInput(sourcePath, destPath, fileNames)
	if err != nil {
		return 0, err
	}

	copied := 0
	for _, fileName := range fileNames {
		if !isSafeFileName(fileName) {
			return copied, fmt.Errorf("ชื่อไฟล์ไม่ถูกต้อง: %s", fileName)
		}
		sourceFile := filepath.Join(sourceAbsolute, fileName)
		destFile := filepath.Join(destAbsolute, fileName)
		if err := copyOneFile(sourceFile, destFile); err != nil {
			return copied, fmt.Errorf("คัดลอกไฟล์ %s ไม่สำเร็จ: %w", fileName, err)
		}
		record := model.BackupRecord{SourcePath: sourceFile, DestinationPath: destAbsolute, FileName: fileName, Status: "copied"}
		ctx, cancel := databaseContext(s.context())
		err = s.repo.Create(ctx, &record)
		cancel()
		if err != nil {
			_ = os.Remove(destFile)
			return copied, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %w", fileName, err)
		}
		copied++
	}
	return copied, nil
}

func (s *backupService) MoveSelectedFiles(sourcePath, destPath string, fileNames []string) (int, error) {
	sourceAbsolute, destAbsolute, err := validateSelectedInput(sourcePath, destPath, fileNames)
	if err != nil {
		return 0, err
	}

	moved := 0
	for _, fileName := range fileNames {
		if !isSafeFileName(fileName) {
			return moved, fmt.Errorf("ชื่อไฟล์ไม่ถูกต้อง: %s", fileName)
		}
		sourceFile := filepath.Join(sourceAbsolute, fileName)
		destFile := filepath.Join(destAbsolute, fileName)
		if err := requireRegularFile(sourceFile, "ต้นทาง"); err != nil {
			return moved, err
		}
		if _, err := os.Lstat(destFile); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return moved, fmt.Errorf("ตรวจไฟล์ปลายทาง %s ไม่ได้: %w", fileName, err)
		}
		if err := s.moveOneFile(sourceFile, destFile); err != nil {
			return moved, fmt.Errorf("ย้ายไฟล์ %s ไม่สำเร็จ: %w", fileName, err)
		}
		record := model.BackupRecord{SourcePath: sourceFile, DestinationPath: destAbsolute, FileName: fileName, Status: "moved"}
		ctx, cancel := databaseContext(s.context())
		err = s.repo.Create(ctx, &record)
		cancel()
		if err != nil {
			if rollbackErr := s.moveOneFile(destFile, sourceFile); rollbackErr != nil {
				return moved, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %v; และย้ายไฟล์กลับต้นทางไม่ได้: %w", fileName, err, rollbackErr)
			}
			return moved, fmt.Errorf("บันทึกประวัติไฟล์ %s ไม่สำเร็จ: %w", fileName, err)
		}
		moved++
	}
	return moved, nil
}

func (s *backupService) DeleteSelectedFiles(destPath string, fileNames []string) (int, error) {
	if len(fileNames) == 0 {
		return 0, fmt.Errorf("กรุณาเลือกไฟล์ก่อน")
	}
	destAbsolute, err := filepath.Abs(strings.TrimSpace(destPath))
	if err != nil || strings.TrimSpace(destPath) == "" {
		return 0, fmt.Errorf("กรุณาตั้งค่าโฟลเดอร์ปลายทางก่อน")
	}
	destInfo, err := os.Stat(destAbsolute)
	if err != nil {
		return 0, fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}
	if !destInfo.IsDir() {
		return 0, fmt.Errorf("ปลายทางไม่ใช่โฟลเดอร์")
	}

	deleted := 0
	for _, fileName := range fileNames {
		if !isSafeFileName(fileName) {
			return deleted, fmt.Errorf("ชื่อไฟล์ไม่ถูกต้อง: %s", fileName)
		}
		filePath := filepath.Join(destAbsolute, fileName)
		if err := requireRegularFile(filePath, "ปลายทาง"); err != nil {
			return deleted, err
		}
		if err := os.Remove(filePath); err != nil {
			return deleted, fmt.Errorf("ลบไฟล์ %s ไม่สำเร็จ: %w", fileName, err)
		}
		deleted++
	}
	return deleted, nil
}

func (s *backupService) GetHistory(destinationPath string) ([]model.BackupRecord, error) {
	destination, err := absoluteFolderPath(destinationPath, "ปลายทาง")
	if err != nil {
		return nil, err
	}
	ctx, cancel := databaseContext(s.context())
	defer cancel()
	return s.repo.ListByDestination(ctx, destination)
}

func (s *backupService) FindMissingFiles(destinationPath string) ([]string, error) {
	destination, err := absoluteFolderPath(destinationPath, "ปลายทาง")
	if err != nil {
		return nil, err
	}
	ctx, cancel := databaseContext(s.context())
	records, err := s.repo.ListMissingCandidates(ctx, destination)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("อ่านประวัติไม่ได้: %w", err)
	}

	missing := make([]string, 0)
	for _, record := range records {
		filePath := filepath.Join(destination, record.FileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			missing = append(missing, record.FileName)
		} else if err != nil {
			return nil, fmt.Errorf("ตรวจไฟล์ %s ไม่ได้: %w", record.FileName, err)
		}
	}
	return missing, nil
}

func validateSelectedInput(sourcePath, destPath string, fileNames []string) (string, string, error) {
	if len(fileNames) == 0 {
		return "", "", fmt.Errorf("กรุณาเลือกไฟล์ก่อน")
	}
	return validateFolders(sourcePath, destPath)
}

func validateFolders(sourcePath, destPath string) (string, string, error) {
	source, err := absoluteFolderPath(sourcePath, "ต้นทาง")
	if err != nil {
		return "", "", err
	}
	destination, err := absoluteFolderPath(destPath, "ปลายทาง")
	if err != nil {
		return "", "", err
	}
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return "", "", fmt.Errorf("ตรวจสอบโฟลเดอร์ต้นทางไม่ได้: %w", err)
	}
	destinationInfo, err := os.Stat(destination)
	if err != nil {
		return "", "", fmt.Errorf("ตรวจสอบโฟลเดอร์ปลายทางไม่ได้: %w", err)
	}
	if !sourceInfo.IsDir() {
		return "", "", fmt.Errorf("ต้นทางไม่ใช่โฟลเดอร์")
	}
	if !destinationInfo.IsDir() {
		return "", "", fmt.Errorf("ปลายทางไม่ใช่โฟลเดอร์")
	}
	if os.SameFile(sourceInfo, destinationInfo) {
		return "", "", fmt.Errorf("โฟลเดอร์ต้นทางกับปลายทางต้องไม่ใช่โฟลเดอร์เดียวกัน")
	}
	return source, destination, nil
}

func absoluteFolderPath(path, label string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("กรุณาระบุโฟลเดอร์%s", label)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("ตรวจสอบโฟลเดอร์%sไม่ได้: %w", label, err)
	}
	return absolute, nil
}

func isSafeFileName(fileName string) bool {
	return fileName != "" && fileName != "." && fileName != ".." && filepath.Base(fileName) == fileName && !strings.ContainsAny(fileName, `/\\`)
}

func requireRegularFile(path, location string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("ตรวจไฟล์%sไม่ได้: %w", location, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("ไฟล์%sไม่ใช่ไฟล์ปกติ", location)
	}
	return nil
}

func copyOneFile(sourcePath, destPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return fmt.Errorf("ตรวจไฟล์ต้นทางไม่ได้: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("ไฟล์ต้นทางไม่ใช่ไฟล์ปกติ")
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("เปิดไฟล์ต้นทางไม่ได้: %w", err)
	}
	openedInfo, err := source.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		_ = source.Close()
		if err != nil {
			return fmt.Errorf("ตรวจไฟล์ต้นทางไม่ได้: %w", err)
		}
		return fmt.Errorf("ไฟล์ต้นทางเปลี่ยนระหว่างทำงาน")
	}
	destination, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		_ = source.Close()
		return fmt.Errorf("สร้างไฟล์ปลายทางไม่ได้: %w", err)
	}
	_, copyErr := io.Copy(destination, source)
	sourceCloseErr := source.Close()
	destinationCloseErr := destination.Close()
	if copyErr != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("คัดลอกข้อมูลไม่สำเร็จ: %w", copyErr)
	}
	if sourceCloseErr != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("ปิดไฟล์ต้นทางไม่ได้: %w", sourceCloseErr)
	}
	if destinationCloseErr != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("ปิดไฟล์ปลายทางไม่ได้: %w", destinationCloseErr)
	}
	return nil
}

func (s *backupService) moveOneFile(sourcePath, destPath string) error {
	if _, err := os.Lstat(destPath); err == nil {
		return fmt.Errorf("ไฟล์ปลายทางมีอยู่แล้ว: %s", destPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("ตรวจไฟล์ปลายทางไม่ได้: %w", err)
	}
	if err := os.Rename(sourcePath, destPath); err == nil {
		return nil
	}
	if err := copyOneFile(sourcePath, destPath); err != nil {
		return err
	}
	if err := os.Remove(sourcePath); err != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("ลบไฟล์ต้นทางไม่สำเร็จ: %w", err)
	}
	return nil
}
