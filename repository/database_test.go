package repository

import (
	"path/filepath"
	"testing"
)

func TestBackupTable(t *testing.T) {
	testDatabasePath := filepath.Join(t.TempDir(), "backup_test.db")
	t.Setenv("DATABASE", testDatabasePath)

	db, err := NewDbConnection()
	if err != nil {
		t.Fatalf("เปิดฐานข้อมูลไม่สำเร็จ: %v", err)
	}

	if !db.Migrator().HasTable("backup_records") {
		t.Fatal("ไม่พบตาราง backup_records")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("อ่านฐานข้อมูลไม่สำเร็จ: %v", err)
	}
	defer sqlDB.Close()
}
