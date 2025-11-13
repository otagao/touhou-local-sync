package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicCopy performs an atomic file copy operation.
// It writes to a temporary file first, then atomically renames it to the destination.
// This prevents partial writes in case of errors.
//
// Steps:
// 1. Create a .tmp file in the same directory as dest
// 2. Copy src to .tmp
// 3. Atomically rename .tmp to dest
// 4. If any error occurs, clean up the .tmp file
func AtomicCopy(src, dest string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create temporary file in the same directory as destination
	destDir := filepath.Dir(dest)
	tmpFile, err := os.CreateTemp(destDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Copy data
	if _, err = io.Copy(tmpFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Sync to ensure data is written to disk
	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file before rename
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions to match source
	if err = os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err = os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// FileExists checks if a file exists and is readable.
func FileExists(path string) (exists bool, readable bool) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false
		}
		// File exists but stat failed (possibly permission issue)
		return true, false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return true, false
	}

	// Try to open for read to verify readability
	file, err := os.Open(path)
	if err != nil {
		return true, false
	}
	file.Close()

	return true, true
}

// ExpandEnvPath expands environment variables in a path (e.g., %APPDATA%)
func ExpandEnvPath(path string) string {
	return os.ExpandEnv(path)
}
