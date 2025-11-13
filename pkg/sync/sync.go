package sync

import (
	"fmt"
	"path/filepath"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/backup"
	"github.com/otagao/touhou-local-sync/pkg/process"
	"github.com/otagao/touhou-local-sync/pkg/utils"
)

// PullFile synchronizes a file from local to USB (vault).
// This is the "pull" operation - pulling local changes to the central vault.
//
// Steps:
// 1. Compare local and vault files
// 2. If local is preferred, backup vault file
// 3. Copy local to vault atomically
func PullFile(title string, localPath string, vaultPath string) (*models.ComparisonResult, error) {
	// Get metadata for both files
	localMeta, err := GetFileMetadata(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get local metadata: %w", err)
	}

	vaultMeta, err := GetFileMetadata(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault metadata: %w", err)
	}

	// Compare files
	comparison := CompareFiles(localMeta, vaultMeta)

	// Only proceed if recommendation is PULL
	if comparison.Recommendation != "PULL" {
		return comparison, nil
	}

	// Ensure vault directory exists
	vaultDir := filepath.Dir(vaultPath)
	if err := utils.EnsureDir(vaultDir); err != nil {
		return comparison, fmt.Errorf("failed to create vault directory: %w", err)
	}

	// Backup existing vault file if it exists
	if vaultMeta.Exists && vaultMeta.Readable {
		_, err := backup.CreateBackup(title, vaultPath)
		if err != nil {
			return comparison, fmt.Errorf("failed to backup vault file: %w", err)
		}
	}

	// Copy local to vault
	if err := utils.AtomicCopy(localPath, vaultPath); err != nil {
		return comparison, fmt.Errorf("failed to copy file: %w", err)
	}

	return comparison, nil
}

// PushFile synchronizes a file from USB (vault) to local.
// This is the "push" operation - pushing vault changes to local machines.
//
// Steps:
// 1. Check if local file is safe to write (no game running, not locked)
// 2. Compare vault and local files
// 3. If vault is preferred, backup local file
// 4. Copy vault to local atomically
func PushFile(title string, vaultPath string, localPath string, force bool) (*models.ComparisonResult, error) {
	// Check if it's safe to write to local file
	safe, reason, err := process.CanSafelyWrite(localPath, title)
	if err != nil {
		return nil, fmt.Errorf("failed to check if safe to write: %w", err)
	}
	if !safe && !force {
		return nil, fmt.Errorf("cannot push: %s (use --force to override)", reason)
	}

	// Get metadata for both files
	vaultMeta, err := GetFileMetadata(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault metadata: %w", err)
	}

	localMeta, err := GetFileMetadata(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get local metadata: %w", err)
	}

	// Compare files
	comparison := CompareFiles(localMeta, vaultMeta)

	// Only proceed if recommendation is PUSH
	if comparison.Recommendation != "PUSH" {
		// If local is newer, warn and skip unless forced
		if comparison.Recommendation == "PULL" && !force {
			return comparison, fmt.Errorf("local file appears newer than vault, skipping push (use --force to override)")
		}
		if comparison.Recommendation == "SKIP" {
			return comparison, nil
		}
		if comparison.Recommendation == "CONFLICT" && !force {
			return comparison, fmt.Errorf("file conflict detected: %s (use --force to override)", comparison.Reason)
		}
	}

	// Ensure local directory exists
	localDir := filepath.Dir(localPath)
	if err := utils.EnsureDir(localDir); err != nil {
		return comparison, fmt.Errorf("failed to create local directory: %w", err)
	}

	// Backup existing local file if it exists
	if localMeta.Exists && localMeta.Readable {
		_, err := backup.CreateBackup(title, localPath)
		if err != nil {
			return comparison, fmt.Errorf("failed to backup local file: %w", err)
		}
	}

	// Copy vault to local
	if err := utils.AtomicCopy(vaultPath, localPath); err != nil {
		return comparison, fmt.Errorf("failed to copy file: %w", err)
	}

	return comparison, nil
}

// GetPreferredLocalPath returns the preferred local path for a title and device.
// Returns the path from the paths.json configuration.
func GetPreferredLocalPath(pathsConfig *models.PathsConfig, title string, deviceID string) (string, error) {
	// Check if title exists in config
	titlePaths, ok := pathsConfig.Paths[title]
	if !ok {
		return "", fmt.Errorf("no paths configured for title: %s", title)
	}

	// Check if device has paths for this title
	pathEntry, ok := titlePaths[deviceID]
	if !ok {
		return "", fmt.Errorf("no paths configured for device %s on title %s", deviceID, title)
	}

	// Check if paths array is empty
	if len(pathEntry.Paths) == 0 {
		return "", fmt.Errorf("paths array is empty for device %s on title %s", deviceID, title)
	}

	// Check if preferred index is valid
	if pathEntry.Preferred < 0 || pathEntry.Preferred >= len(pathEntry.Paths) {
		return "", fmt.Errorf("invalid preferred index %d for device %s on title %s", pathEntry.Preferred, deviceID, title)
	}

	// Get preferred path and expand environment variables
	path := pathEntry.Paths[pathEntry.Preferred]
	expandedPath := utils.ExpandEnvPath(path)

	return expandedPath, nil
}

// GetVaultFilePath returns the vault file path for a title.
// Example: <vault>/th08/main/score.dat
func GetVaultFilePath(title string, filename string) (string, error) {
	vaultPath, err := backup.GetTitleVaultPath(title)
	if err != nil {
		return "", err
	}

	return filepath.Join(vaultPath, filename), nil
}
