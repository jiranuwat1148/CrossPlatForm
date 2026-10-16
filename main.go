package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const databaseFile = "crossplatform.db"

// File maps to the supplied database table:
// files(idx, dest, filename)
type File struct {
	Idx      uint   `gorm:"column:idx;primaryKey;autoIncrement"`
	Dest     string `gorm:"column:dest;not null"`
	Filename string `gorm:"column:filename;not null"`
}

func (File) TableName() string {
	return "files"
}

// Setting maps to the supplied database table:
// settings(idx, source, dest)
type Setting struct {
	Idx    uint   `gorm:"column:idx;primaryKey;autoIncrement"`
	Source string `gorm:"column:source;not null"`
	Dest   string `gorm:"column:dest;not null"`
}

func (Setting) TableName() string {
	return "settings"
}

var (
	db         *gorm.DB
	sourcePath string
	destPath   string
)

func main() {
	if err := initDB(); err != nil {
		fmt.Println("[FATAL] Cannot initialize database:", err)
		return
	}

	loadLastSetting()

	fmt.Println("=== Flash Drive Backup CLI ===")
	fmt.Println("Type /help to show commands")

	if destPath == "" {
		fmt.Println("[WARNING] Destination is not set! Please run '/dest <path>' first.")
	} else {
		fmt.Println("[OK] Destination already set to:", destPath)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			fmt.Println("Program terminated.")
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if !handleCommand(input) {
			return
		}
	}
}

// initDB opens SQLite through a pure-Go driver.
// GORM logging is silent so SQL/debug output does not clutter the exam output.
func initDB() error {
	var err error

	db, err = gorm.Open(
		sqlite.Open(databaseFile),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		return err
	}

	// Ensures required tables exist.
	if err := db.AutoMigrate(&File{}, &Setting{}); err != nil {
		return err
	}

	return nil
}

// loadLastSetting checks the latest saved Source/Destination when the program starts.
// If no setting exists, sourcePath and destPath remain empty.
func loadLastSetting() {
	var setting Setting

	result := db.
		Order("idx DESC").
		Limit(1).
		Find(&setting)

	if result.Error != nil || result.RowsAffected == 0 {
		return
	}

	// Restore Source only if the folder still exists.
	if setting.Source != "" {
		if info, err := os.Stat(setting.Source); err == nil && info.IsDir() {
			sourcePath = filepath.Clean(setting.Source)
		}
	}

	// Restore Destination only if the folder still exists.
	if setting.Dest != "" {
		if info, err := os.Stat(setting.Dest); err == nil && info.IsDir() {
			destPath = filepath.Clean(setting.Dest)
		}
	}
}

func handleCommand(input string) bool {
	command, args := splitCommand(input)

	switch strings.ToLower(command) {
	case "/source":
		handleSource(args)

	case "/dest":
		handleDest(args)

	case "/list":
		handleList(args)

	case "/move":
		handleMove(args)

	case "/check":
		handleCheck()

	case "/delete":
		handleDelete(args)

	case "/help":
		showHelp()

	case "/exit":
		fmt.Println("Program terminated. Good bye!")
		return false

	default:
		fmt.Println("[ERROR] Unknown command.")
	}

	return true
}

func splitCommand(input string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(input), " ", 2)

	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], strings.TrimSpace(parts[1])
}


func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  /source <path>                  Set source folder")
	fmt.Println("  /dest <path>                    Set destination folder")
	fmt.Println("  /list source                    List files in source")
	fmt.Println("  /list dest                      List files in destination")
	fmt.Println("  /list db [path]                 List DB history for destination")
	fmt.Println("  /move <file1>, <file2>, ...     Move selected files")
	fmt.Println("  /move all                       Move all files")
	fmt.Println("  /check                          Check DB records against destination")
	fmt.Println("  /delete dest <filename>         Delete file from destination and DB")
	fmt.Println("  /delete all                     Delete all files in destination")
	fmt.Println("  /help                           Show this help")
	fmt.Println("  /exit                           Exit program")
}

// Step 2: set source path.
func handleSource(args string) {
	path := cleanPath(args)

	if path == "" {
		fmt.Println("[ERROR] Usage: /source <path>")
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("[ERROR] Source path not found.")
		return
	}

	if !info.IsDir() {
		fmt.Println("[ERROR] Source must be a directory.")
		return
	}

	sourcePath = filepath.Clean(path)

	if err := saveSetting(); err != nil {
		fmt.Println("[DB ERROR] Cannot save source/destination setting.")
		return
	}

	count, err := countFiles(sourcePath)
	if err != nil {
		fmt.Println("[ERROR] Cannot read source directory.")
		return
	}

	fmt.Printf("[OK] Source set to: %s (Found %d files)\n", sourcePath, count)
}

// Step 2 + Step 6: set destination, save through GORM, then integrity scan.
func handleDest(args string) {
	path := cleanPath(args)

	if path == "" {
		fmt.Println("[ERROR] Usage: /dest <path>")
		return
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Println("[ERROR] Cannot create/access destination.")
		return
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		fmt.Println("[ERROR] Destination must be a directory.")
		return
	}

	destPath = filepath.Clean(path)

	if err := saveSetting(); err != nil {
		fmt.Println("[DB ERROR] Cannot save source/destination setting.")
		return
	}

	fmt.Println("[OK] Destination set to:", destPath)
	checkIntegrityFor(destPath, true)
}

// saveSetting writes the current state through GORM.
// Because the supplied settings schema requires BOTH source and dest,
// a record is inserted only after both are available.
func saveSetting() error {
	if sourcePath == "" || destPath == "" {
		return nil
	}

	// Avoid adding the exact same setting repeatedly.
	var last Setting
	result := db.
		Order("idx DESC").
		Limit(1).
		Find(&last)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 &&
		last.Source == sourcePath &&
		last.Dest == destPath {
		return nil
	}

	return db.Create(&Setting{
		Source: sourcePath,
		Dest:   destPath,
	}).Error
}

func countFiles(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count, nil
}

// Step 1 + Step 5.
func handleList(args string) {
	target, extra := splitCommand(args)

	switch strings.ToLower(target) {
	case "source":
		listDirectory(sourcePath, "Source")

	case "dest":
		listDirectory(destPath, "Destination")

	case "db":
		path := cleanPath(extra)

		if path == "" {
			path = destPath
		}

		// Step 1 State Guard.
		if path == "" {
			fmt.Println("[ERROR] Destination is not set! Please run '/dest <path>' or '/list db <path>'.")
			return
		}

		listDB(filepath.Clean(path))

	default:
		fmt.Println("[ERROR] Usage: /list source | /list dest | /list db [path]")
	}
}

func listDirectory(path, label string) {
	if path == "" {
		fmt.Printf("[ERROR] %s is not set!\n", label)
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("[ERROR] Cannot read directory.")
		return
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fmt.Println(entry.Name())
		count++
	}

	if count == 0 {
		fmt.Println("[INFO] No files.")
	}
}

// Step 5: only show DB history for the requested/current destination.
func listDB(path string) {
	var records []File

	result := db.
		Where("dest = ?", path).
		Order("idx ASC").
		Find(&records)

	if result.Error != nil {
		fmt.Println("[DB ERROR] Cannot query backup history.")
		return
	}

	if len(records) == 0 {
		fmt.Println("[INFO] No backup history for:", path)
		return
	}

	for _, record := range records {
		fmt.Printf("%d | %s\n", record.Idx, record.Filename)
	}
}

// Step 3: move all or selected files, log to DB, show total elapsed time.
func handleMove(args string) {
	if sourcePath == "" {
		fmt.Println("[ERROR] Source is not set! Please run '/source <path>' first.")
		return
	}

	if destPath == "" {
		fmt.Println("[ERROR] Destination is not set! Please run '/dest <path>' first.")
		return
	}

	args = strings.TrimSpace(args)
	if args == "" {
		fmt.Println("[ERROR] Usage: /move <file1>, <file2>, ... OR /move all")
		return
	}

	files, err := resolveSelection(sourcePath, args)
	if err != nil {
		fmt.Println("[ERROR] Cannot read source directory.")
		return
	}

	start := time.Now()

	moved := 0
	skipped := 0
	failed := 0

	for _, fileName := range files {
		if err := validateFilename(fileName); err != nil {
			fmt.Printf("[FAILED] %s: invalid filename\n", fileName)
			failed++
			continue
		}

		src := filepath.Join(sourcePath, fileName)
		dst := filepath.Join(destPath, fileName)

		info, err := os.Stat(src)
		if err != nil || info.IsDir() {
			fmt.Printf("[FAILED] %s: source file not found\n", fileName)
			failed++
			continue
		}

		// Step 4: duplicate protection on BOTH disk and DB.
		duplicate, location, err := isDuplicate(fileName)
		if err != nil {
			fmt.Printf("[FAILED] %s: duplicate check error\n", fileName)
			failed++
			continue
		}

		if duplicate {
			fmt.Printf("[SKIPPED] %s: Duplicate file found in %s!\n", fileName, location)
			skipped++
			continue
		}

		method, err := moveFile(src, dst, info.Mode())
		if err != nil {
			fmt.Printf("[FAILED] %s: %v\n", fileName, err)
			failed++
			continue
		}

		// Log only after the physical move is successful.
		record := File{
			Dest:     destPath,
			Filename: fileName,
		}

		if err := db.Create(&record).Error; err != nil {
			fmt.Printf("[DB ERROR] %s moved but DB logging failed.\n", fileName)
			failed++
			continue
		}

		fmt.Printf("[SUCCESS] %s moved (%s)\n", fileName, method)
		moved++
	}

	elapsed := time.Since(start)

	fmt.Printf(
		"[Done] %d moved, %d skipped, %d failed | %.2f ms (%.3f s)\n",
		moved,
		skipped,
		failed,
		float64(elapsed.Microseconds())/1000.0,
		elapsed.Seconds(),
	)
}

func resolveSelection(directory, args string) ([]string, error) {
	if strings.EqualFold(strings.TrimSpace(args), "all") {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		var files []string

		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}

		return files, nil
	}

	items := strings.Split(args, ",")
	files := make([]string, 0, len(items))

	for _, item := range items {
		name := strings.TrimSpace(strings.Trim(item, `"'`))
		if name != "" {
			files = append(files, name)
		}
	}

	return files, nil
}

// Step 4: check BOTH destination folder and destination-scoped DB.
func isDuplicate(fileName string) (bool, string, error) {
	dst := filepath.Join(destPath, fileName)

	if _, err := os.Stat(dst); err == nil {
		return true, "Destination", nil
	} else if !os.IsNotExist(err) {
		return false, "", err
	}

	var count int64

	result := db.
		Model(&File{}).
		Where("filename = ? AND dest = ?", fileName, destPath).
		Count(&count)

	if result.Error != nil {
		return false, "", result.Error
	}

	if count > 0 {
		return true, "Database", nil
	}

	return false, "", nil
}

// Step 3 cross-device handling:
// 1) try os.Rename
// 2) if it fails, Copy + Sync + Delete source.
func moveFile(src, dst string, mode os.FileMode) (string, error) {
	if err := os.Rename(src, dst); err == nil {
		return "Move", nil
	}

	if err := copyFile(src, dst, mode); err != nil {
		return "", errors.New("move/copy failed")
	}

	if err := os.Remove(src); err != nil {
		// Prevent an incomplete "move" state.
		_ = os.Remove(dst)
		return "", errors.New("copied but source delete failed")
	}

	return "Cross-device Copy+Delete", nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// O_EXCL guarantees an existing destination file is never overwritten.
	destination, err := os.OpenFile(
		dst,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		mode.Perm(),
	)
	if err != nil {
		return err
	}

	success := false

	defer func() {
		_ = destination.Close()

		if !success {
			_ = os.Remove(dst)
		}
	}()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	if err := destination.Sync(); err != nil {
		return err
	}

	if err := destination.Close(); err != nil {
		return err
	}

	success = true
	return nil
}

// Step 6.
func handleCheck() {
	if destPath == "" {
		fmt.Println("[ERROR] Destination is not set! Please run '/dest <path>' first.")
		return
	}

	checkIntegrityFor(destPath, false)
}

func checkIntegrityFor(path string, fromDestCommand bool) {
	var records []File

	result := db.
		Where("dest = ?", path).
		Order("idx ASC").
		Find(&records)

	if result.Error != nil {
		fmt.Println("[DB ERROR] Integrity check failed.")
		return
	}

	if len(records) == 0 {
		if !fromDestCommand {
			fmt.Println("[OK] No recorded files to verify.")
		}
		return
	}

	missing := 0

	for _, record := range records {
		filePath := filepath.Join(path, record.Filename)

		info, err := os.Stat(filePath)

		if os.IsNotExist(err) {
			fmt.Printf("[ALERT] Missing file: %s\n", record.Filename)
			missing++
			continue
		}

		if err != nil || info.IsDir() {
			fmt.Printf("[ALERT] Invalid file: %s\n", record.Filename)
			missing++
		}
	}

	if missing == 0 && !fromDestCommand {
		fmt.Println("[OK] Integrity check passed.")
	}
}

// Step 7.
// Supports the rubric form: /delete dest <filename>
// Also supports /delete all and comma-separated filenames.
func handleDelete(args string) {
	if destPath == "" {
		fmt.Println("[ERROR] Destination is not set! Please run '/dest <path>' first.")
		return
	}

	args = strings.TrimSpace(args)

	if args == "" {
		fmt.Println("[ERROR] Usage: /delete dest <filename> | /delete all")
		return
	}

	lower := strings.ToLower(args)

	if lower == "dest" {
		fmt.Println("[ERROR] Usage: /delete dest <filename>")
		return
	}

	if strings.HasPrefix(lower, "dest ") {
		args = strings.TrimSpace(args[len("dest "):])
	}

	var files []string

	if strings.EqualFold(args, "all") {
		entries, err := os.ReadDir(destPath)
		if err != nil {
			fmt.Println("[ERROR] Cannot read destination directory.")
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
	} else {
		var err error
		files, err = resolveSelection(destPath, args)
		if err != nil {
			fmt.Println("[ERROR] Invalid delete request.")
			return
		}
	}

	deleted := 0
	failed := 0

	for _, fileName := range files {
		if err := validateFilename(fileName); err != nil {
			fmt.Printf("[FAILED] %s: invalid filename\n", fileName)
			failed++
			continue
		}

		filePath := filepath.Join(destPath, fileName)

		if err := os.Remove(filePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("[FAILED] %s: file not found\n", fileName)
			} else {
				fmt.Printf("[FAILED] %s: cannot delete file\n", fileName)
			}

			failed++
			continue
		}

		// The supplied DB has no status column.
		// Therefore we synchronize DB state by deleting matching history through GORM.
		result := db.
			Where("filename = ? AND dest = ?", fileName, destPath).
			Delete(&File{})

		if result.Error != nil {
			fmt.Printf("[DB ERROR] %s: database update failed\n", fileName)
			failed++
			continue
		}

		fmt.Printf("[SUCCESS] %s deleted and database updated.\n", fileName)
		deleted++
	}

	if len(files) > 1 {
		fmt.Printf("[Done] %d deleted, %d failed.\n", deleted, failed)
	}
}

func validateFilename(name string) error {
	if name == "" {
		return errors.New("empty filename")
	}

	if filepath.Base(name) != name {
		return errors.New("nested path is not allowed")
	}

	if name == "." || name == ".." {
		return errors.New("invalid filename")
	}

	return nil
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)

	if len(path) >= 2 {
		if (path[0] == '"' && path[len(path)-1] == '"') ||
			(path[0] == '\'' && path[len(path)-1] == '\'') {
			path = path[1 : len(path)-1]
		}
	}

	return strings.TrimSpace(path)
}
