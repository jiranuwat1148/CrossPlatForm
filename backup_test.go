package main

import (
	"Backup_image/model"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStartupDB(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "backup.db")
	t.Setenv("DATABASE", databasePath)

	app := NewApp()
	app.startup(context.Background())

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("ฐานข้อมูลไม่ถูกสร้าง: %v", err)
	}
}
func TestMoveFile(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	sourceFile := filepath.Join(sourceDir, "photo.txt")
	destFile := filepath.Join(destDir, "photo.txt")

	if err := os.WriteFile(sourceFile, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := moveOneFile(sourceFile, destFile); err != nil {
		t.Fatalf("ย้ายไฟล์ไม่สำเร็จ: %v", err)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("อ่านไฟล์ปลายทางไม่ได้: %v", err)
	}
	if string(content) != "hello" {
		t.Fatal("ข้อมูลในไฟล์ปลายทางไม่ตรง")
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Fatal("ไฟล์ต้นทางยังอยู่หลังย้าย")
	}
}
func TestMoveFiles(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "photo.txt")

	if err := os.WriteFile(sourceFile, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DATABASE", filepath.Join(t.TempDir(), "backup.db"))

	app := NewApp()
	app.startup(context.Background())

	moved, err := app.MoveFiles(sourceDir, destDir)
	if err != nil {
		t.Fatalf("ย้ายไฟล์ไม่สำเร็จ: %v", err)
	}
	if moved != 1 {
		t.Fatalf("อยากให้ย้ายได้ 1 ไฟล์ แต่ย้ายได้ %d", moved)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "photo.txt"))
	if err != nil {
		t.Fatalf("อ่านไฟล์ปลายทางไม่ได้: %v", err)
	}
	if string(content) != "hello" {
		t.Fatal("ข้อมูลในไฟล์ปลายทางไม่ตรง")
	}

	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Fatal("ไฟล์ต้นทางยังอยู่หลังย้าย")
	}

	var record model.BackupRecord
	err = app.db.
		Where("destination_path = ? AND file_name = ?", destDir, "photo.txt").
		First(&record).Error
	if err != nil {
		t.Fatalf("ไม่พบประวัติไฟล์ในฐานข้อมูล: %v", err)
	}
	destFile := filepath.Join(destDir, "photo.txt")

	if err := os.Remove(destFile); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(sourceFile, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	moved, err = app.MoveFiles(sourceDir, destDir)
	if err != nil {
		t.Fatalf("ย้ายไฟล์ซ้ำไม่สำเร็จ: %v", err)
	}

	if moved != 0 {
		t.Fatalf("ไฟล์ซ้ำควรถูกข้าม แต่ย้ายได้ %d ไฟล์", moved)
	}

	if _, err := os.Stat(sourceFile); err != nil {
		t.Fatalf("ไฟล์ต้นทางควรยังอยู่: %v", err)
	}
	records, err := app.GetHistory(destDir)
	if err != nil {
		t.Fatalf("อ่านประวัติไม่ได้: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("ควรมีประวัติ 1 รายการ แต่ได้ %d", len(records))
	}
	if records[0].FileName != "photo.txt" {
		t.Fatalf("ชื่อไฟล์ในประวัติไม่ตรง: %s", records[0].FileName)
	}
	missingFiles, err := app.FindMissingFiles(destDir)
	if err != nil {
		t.Fatalf("ตรวจไฟล์ไม่สำเร็จ: %v", err)
	}
	if len(missingFiles) != 1 || missingFiles[0] != "photo.txt" {
		t.Fatalf("ควรพบไฟล์ที่หายไป photo.txt แต่ได้ %v", missingFiles)
	}
}
