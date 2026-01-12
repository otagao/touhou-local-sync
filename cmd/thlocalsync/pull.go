package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/backup"
	"github.com/otagao/touhou-local-sync/pkg/config"
	"github.com/otagao/touhou-local-sync/pkg/device"
	"github.com/otagao/touhou-local-sync/pkg/logger"
	"github.com/otagao/touhou-local-sync/pkg/pathdetect"
	"github.com/otagao/touhou-local-sync/pkg/sync"
	"github.com/otagao/touhou-local-sync/pkg/utils"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [title|all]",
	Short: "ローカル → ポータブルストレージ（正本へ吸い上げ）",
	Long: `ローカルのセーブデータをポータブルストレージの正本へ吸い上げます。

ローカルがポータブルストレージより新しい/大きい場合に上書きします。
上書き前にポータブルストレージ側のファイルはバックアップされます。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPull,
}

func runPull(cmd *cobra.Command, args []string) error {
	// Determine target title
	targetTitle := "all"
	if len(args) > 0 {
		targetTitle = args[0]
	}

	// Get device ID
	deviceID, _, hostname, err := device.GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	fmt.Printf("=== thlocalsync pull ===\n")
	fmt.Printf("Device: %s (%s)\n\n", deviceID, hostname)

	// Initialize logger
	log, err := logger.New()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Load configurations
	pathsConfig, err := config.LoadPaths()
	if err != nil {
		return fmt.Errorf("failed to load paths config: %w", err)
	}

	// Get titles to pull
	var titles []string
	if targetTitle == "all" {
		// Get all titles from config
		for title := range pathsConfig.Paths {
			titles = append(titles, title)
		}
		if len(titles) == 0 {
			fmt.Println("No titles configured. Run 'thlocalsync detect' first.")
			return nil
		}
		// Sort by release order
		titles = pathdetect.SortTitlesByRelease(titles)
	} else {
		// Validate title code
		if !pathdetect.IsValidTitleCode(targetTitle) {
			return fmt.Errorf("invalid title code: %s", targetTitle)
		}
		titles = []string{targetTitle}
	}

	// Pull each title
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, title := range titles {
		err := pullTitle(title, deviceID, pathsConfig, log)
		if err != nil {
			fmt.Printf("✗ %s: %v\n", title, err)
			errorCount++
			// Log error
			log.Error("pull_error", map[string]interface{}{
				"title":  title,
				"device": deviceID,
				"error":  err.Error(),
			})
		} else {
			// Check if actually pulled or skipped
			// We'll track this in pullTitle
			successCount++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Success: %d, Skipped: %d, Errors: %d\n", successCount, skipCount, errorCount)

	return nil
}

func pullTitle(title, deviceID string, pathsConfig *models.PathsConfig, log *logger.Logger) error {
	// Get local path
	localPath, err := sync.GetPreferredLocalPath(pathsConfig, title, deviceID)
	if err != nil {
		return fmt.Errorf("no path configured")
	}

	// Determine vault file name
	titleInfo := pathdetect.GetTitleByCode(title)
	var fileName string
	if titleInfo != nil {
		fileName = titleInfo.FileName
	} else {
		fileName = "score.dat"
	}

	// Get vault path
	vaultPath, err := sync.GetVaultFilePath(title, fileName)
	if err != nil {
		return fmt.Errorf("failed to get vault path: %w", err)
	}

	// Pull file
	comparison, err := sync.PullFile(title, localPath, vaultPath)
	if err != nil {
		return err
	}

	// Handle CONFLICT - ask user for resolution
	if comparison.Recommendation == "CONFLICT" {
		choice := promptUserForConflictResolution(title, comparison, "pull")
		switch choice {
		case "local":
			// User chose local - force pull
			comparison, err = sync.ForcePullFile(title, localPath, vaultPath)
			if err != nil {
				return fmt.Errorf("failed to force pull: %w", err)
			}
			fmt.Printf("✓ %s: Pulled to USB (user chose local)\n", title)
			log.Info("pull", map[string]interface{}{
				"title":  title,
				"device": deviceID,
				"action": "update",
				"from":   "local",
				"to":     "usb",
				"reason": "user resolved conflict - chose local",
			})
		case "remote":
			// User chose remote - skip (keep USB version)
			fmt.Printf("- %s: Kept USB version (user choice)\n", title)
			log.Info("pull_skip", map[string]interface{}{
				"title":  title,
				"device": deviceID,
				"reason": "user resolved conflict - chose remote",
			})
		case "cancel":
			fmt.Printf("- %s: Cancelled by user\n", title)
			log.Info("pull_cancel", map[string]interface{}{
				"title":  title,
				"device": deviceID,
				"reason": "user cancelled conflict resolution",
			})
		}
		return nil
	}

	// Report result
	switch comparison.Recommendation {
	case "PULL":
		fmt.Printf("✓ %s: Pulled to USB (%s)\n", title, comparison.Reason)
		// Log operation
		log.Info("pull", map[string]interface{}{
			"title":  title,
			"device": deviceID,
			"action": "update",
			"from":   "local",
			"to":     "usb",
			"reason": comparison.Reason,
		})
	case "SKIP":
		fmt.Printf("- %s: Skipped (%s)\n", title, comparison.Reason)
	case "PUSH":
		fmt.Printf("- %s: USB is newer, skipped (%s)\n", title, comparison.Reason)
	}

	// Archive replays if present
	if err := archiveReplaysIfPresent(title, localPath, log); err != nil {
		log.Error("replay_archive_error", map[string]interface{}{
			"title": title,
			"error": err.Error(),
		})
		// Don't return error - replay archiving is optional
	}

	// Archive snapshots if present
	if err := archiveSnapshotsIfPresent(title, localPath, log); err != nil {
		log.Error("snapshot_archive_error", map[string]interface{}{
			"title": title,
			"error": err.Error(),
		})
		// Don't return error - snapshot archiving is optional
	}

	return nil
}

// hashExistsInArchive checks if a file with the given hash already exists in the archive directory.
func hashExistsInArchive(archiveDir, targetHash string) bool {
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(archiveDir, entry.Name())
		hash, err := utils.CalculateFileHash(filePath)
		if err != nil {
			continue
		}

		if hash == targetHash {
			return true
		}
	}

	return false
}

// archiveReplaysIfPresent archives replay files if the replay directory exists.
func archiveReplaysIfPresent(title, localPath string, log *logger.Logger) error {
	// Detect replay directory
	replayDir := pathdetect.DetectReplayDir(localPath)
	if replayDir == "" {
		log.Info("replay_dir_not_found", map[string]interface{}{
			"title": title,
			"path":  filepath.Join(filepath.Dir(localPath), "replay"),
		})
		return nil
	}

	// List .rpy files
	rpyFiles, err := utils.ListFilesWithExtension(replayDir, ".rpy")
	if err != nil {
		return fmt.Errorf("failed to list replay files: %w", err)
	}

	if len(rpyFiles) == 0 {
		log.Info("no_replay_files", map[string]interface{}{
			"title": title,
		})
		return nil
	}

	// Get archive directory
	archiveDir, err := backup.GetReplayArchiveDir(title)
	if err != nil {
		return fmt.Errorf("failed to get replay archive directory: %w", err)
	}

	// Archive each replay file
	archiveCount := 0
	skipCount := 0

	for _, rpyFile := range rpyFiles {
		srcPath := filepath.Join(replayDir, rpyFile)

		// Calculate hash
		hash, err := utils.CalculateFileHash(srcPath)
		if err != nil {
			log.Warn("replay_hash_failed", map[string]interface{}{
				"title": title,
				"file":  rpyFile,
				"error": err.Error(),
			})
			continue
		}

		// Check if hash already exists in archive
		if hashExistsInArchive(archiveDir, hash) {
			skipCount++
			continue
		}

		// Get file creation time
		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			log.Warn("replay_stat_failed", map[string]interface{}{
				"title": title,
				"file":  rpyFile,
				"error": err.Error(),
			})
			continue
		}
		createdAt := fileInfo.ModTime()

		// Generate archive filename: YYYY-MM-DD_HH-MM-SS_originalname
		archiveName := fmt.Sprintf("%s_%s", createdAt.Format("2006-01-02_15-04-05"), rpyFile)
		archivePath := filepath.Join(archiveDir, archiveName)

		// Atomic copy
		if err := utils.AtomicCopy(srcPath, archivePath); err != nil {
			log.Error("replay_archive_failed", map[string]interface{}{
				"title": title,
				"file":  rpyFile,
				"error": err.Error(),
			})
			continue
		}

		archiveCount++
	}

	// Log summary
	log.Info("replay_archive_complete", map[string]interface{}{
		"title":    title,
		"archived": archiveCount,
		"skipped":  skipCount,
	})

	return nil
}

// archiveSnapshotsIfPresent archives snapshot files if the snapshot directory exists.
func archiveSnapshotsIfPresent(title, localPath string, log *logger.Logger) error {
	// Detect snapshot directory
	snapshotDir := pathdetect.DetectSnapshotDir(localPath)
	if snapshotDir == "" {
		log.Info("snapshot_dir_not_found", map[string]interface{}{
			"title": title,
			"path":  filepath.Join(filepath.Dir(localPath), "snapshot"),
		})
		return nil
	}

	// List .bmp files
	bmpFiles, err := utils.ListFilesWithExtension(snapshotDir, ".bmp")
	if err != nil {
		return fmt.Errorf("failed to list snapshot files: %w", err)
	}

	if len(bmpFiles) == 0 {
		log.Info("no_snapshot_files", map[string]interface{}{
			"title": title,
		})
		return nil
	}

	// Get archive directory
	archiveDir, err := backup.GetSnapshotArchiveDir(title)
	if err != nil {
		return fmt.Errorf("failed to get snapshot archive directory: %w", err)
	}

	// Archive each snapshot file
	archiveCount := 0
	skipCount := 0

	for _, bmpFile := range bmpFiles {
		srcPath := filepath.Join(snapshotDir, bmpFile)

		// Calculate hash
		hash, err := utils.CalculateFileHash(srcPath)
		if err != nil {
			log.Warn("snapshot_hash_failed", map[string]interface{}{
				"title": title,
				"file":  bmpFile,
				"error": err.Error(),
			})
			continue
		}

		// Check if hash already exists in archive
		if hashExistsInArchive(archiveDir, hash) {
			skipCount++
			continue
		}

		// Get file creation time
		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			log.Warn("snapshot_stat_failed", map[string]interface{}{
				"title": title,
				"file":  bmpFile,
				"error": err.Error(),
			})
			continue
		}
		createdAt := fileInfo.ModTime()

		// Generate archive filename: YYYY-MM-DD_HH-MM-SS_originalname
		archiveName := fmt.Sprintf("%s_%s", createdAt.Format("2006-01-02_15-04-05"), bmpFile)
		archivePath := filepath.Join(archiveDir, archiveName)

		// Atomic copy
		if err := utils.AtomicCopy(srcPath, archivePath); err != nil {
			log.Error("snapshot_archive_failed", map[string]interface{}{
				"title": title,
				"file":  bmpFile,
				"error": err.Error(),
			})
			continue
		}

		archiveCount++
	}

	// Log summary
	log.Info("snapshot_archive_complete", map[string]interface{}{
		"title":    title,
		"archived": archiveCount,
		"skipped":  skipCount,
	})

	return nil
}
