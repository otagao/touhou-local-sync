// Package sync handles file synchronization logic and comparison.
package sync

import (
	"fmt"
	"os"

	"github.com/smelt02/touhou-local-sync/internal/models"
	"github.com/smelt02/touhou-local-sync/pkg/utils"
)

// GetFileMetadata retrieves metadata for a file.
// Returns nil if the file doesn't exist or can't be read.
func GetFileMetadata(path string) (*models.FileMetadata, error) {
	meta := &models.FileMetadata{
		Path: path,
	}

	// Check existence and readability
	exists, readable := utils.FileExists(path)
	meta.Exists = exists
	meta.Readable = readable

	if !exists {
		return meta, nil
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return meta, fmt.Errorf("failed to stat file: %w", err)
	}

	meta.Size = info.Size()
	meta.ModTime = info.ModTime().UTC()

	// Calculate hash if readable
	if readable {
		hash, err := utils.CalculateFileHash(path)
		if err != nil {
			return meta, fmt.Errorf("failed to calculate hash: %w", err)
		}
		meta.Hash = hash
	}

	return meta, nil
}
