// Package backup handles history management for save files.
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/otagao/touhou-local-sync/pkg/utils"
)

const (
	// HistoryDir is the subdirectory name for history backups
	HistoryDir = "_history"
)

// GetVaultDir returns the path to the vault directory.
// Assumes vault is at <exe_dir>/vault
func GetVaultDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "vault"), nil
}

// GetTitleVaultPath returns the path to a title's vault directory.
// Example: <vault>/th08/main
func GetTitleVaultPath(title string) (string, error) {
	vaultDir, err := GetVaultDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(vaultDir, title, "main"), nil
}

// GetHistoryDir returns the path to a title's history directory.
// Example: <vault>/th08/_history
func GetHistoryDir(title string) (string, error) {
	vaultDir, err := GetVaultDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(vaultDir, title, HistoryDir), nil
}

// CreateBackup creates a backup of the specified file in the history directory.
// Returns the path to the created backup file.
func CreateBackup(title string, sourceFile string) (string, error) {
	historyDir, err := GetHistoryDir(title)
	if err != nil {
		return "", err
	}

	// Ensure history directory exists
	if err := utils.EnsureDir(historyDir); err != nil {
		return "", fmt.Errorf("failed to create history directory: %w", err)
	}

	// Check if source file exists
	exists, readable := utils.FileExists(sourceFile)
	if !exists {
		return "", fmt.Errorf("source file does not exist: %s", sourceFile)
	}
	if !readable {
		return "", fmt.Errorf("source file is not readable: %s", sourceFile)
	}

	// Generate backup filename with ISO8601 timestamp
	// Format: 2025-11-11T06-20-30Z-score.dat
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	sourceBaseName := filepath.Base(sourceFile)
	backupName := fmt.Sprintf("%s-%s", timestamp, sourceBaseName)
	backupPath := filepath.Join(historyDir, backupName)

	// Copy file to history
	if err := utils.AtomicCopy(sourceFile, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// ListBackups returns a list of backup files for a title, sorted by timestamp (newest first).
func ListBackups(title string) ([]string, error) {
	historyDir, err := GetHistoryDir(title)
	if err != nil {
		return nil, err
	}

	// Check if history directory exists
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	// Collect backup files
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		backups = append(backups, entry.Name())
	}

	// Sort by name (which includes timestamp) in descending order
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	return backups, nil
}

// RestoreBackup restores a backup file to the vault main directory.
// backupName should be the filename only (e.g., "2025-11-11T06-20-30Z-score.dat")
func RestoreBackup(title string, backupName string, targetFile string) error {
	historyDir, err := GetHistoryDir(title)
	if err != nil {
		return err
	}

	backupPath := filepath.Join(historyDir, backupName)

	// Check if backup exists
	exists, readable := utils.FileExists(backupPath)
	if !exists {
		return fmt.Errorf("backup file does not exist: %s", backupName)
	}
	if !readable {
		return fmt.Errorf("backup file is not readable: %s", backupName)
	}

	// Before restoring, create a backup of the current target file if it exists
	if targetExists, _ := utils.FileExists(targetFile); targetExists {
		if _, err := CreateBackup(title, targetFile); err != nil {
			return fmt.Errorf("failed to backup current file before restore: %w", err)
		}
	}

	// Copy backup to target
	if err := utils.AtomicCopy(backupPath, targetFile); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// CleanupOldBackups removes old backups beyond the history limit.
func CleanupOldBackups(title string, limit int) error {
	backups, err := ListBackups(title)
	if err != nil {
		return err
	}

	// If we're under the limit, nothing to do
	if len(backups) <= limit {
		return nil
	}

	historyDir, err := GetHistoryDir(title)
	if err != nil {
		return err
	}

	// Remove backups beyond the limit
	for i := limit; i < len(backups); i++ {
		backupPath := filepath.Join(historyDir, backups[i])
		if err := os.Remove(backupPath); err != nil {
			return fmt.Errorf("failed to remove old backup %s: %w", backups[i], err)
		}
	}

	return nil
}

// GetBackupInfo returns formatted information about a backup file.
type BackupInfo struct {
	Name      string
	Path      string
	Timestamp time.Time
	Size      int64
	Error     error
}

// GetBackupDetails returns detailed information about backups.
func GetBackupDetails(title string) ([]BackupInfo, error) {
	backups, err := ListBackups(title)
	if err != nil {
		return nil, err
	}

	historyDir, err := GetHistoryDir(title)
	if err != nil {
		return nil, err
	}

	var details []BackupInfo
	for _, backup := range backups {
		backupPath := filepath.Join(historyDir, backup)

		info := BackupInfo{
			Name: backup,
			Path: backupPath,
		}

		// Parse timestamp from filename (format: 2025-11-11T06-20-30Z-score.dat)
		parts := strings.Split(backup, "-")
		if len(parts) >= 6 {
			// Reconstruct timestamp string
			timestampStr := strings.Join(parts[:6], "-")
			timestampStr = strings.Replace(timestampStr, "-", ":", 2) // Fix time colons
			if t, err := time.Parse("2006-01-02T15:04:05Z", timestampStr); err == nil {
				info.Timestamp = t
			}
		}

		// Get file size
		if stat, err := os.Stat(backupPath); err == nil {
			info.Size = stat.Size()
		} else {
			info.Error = err
		}

		details = append(details, info)
	}

	return details, nil
}
