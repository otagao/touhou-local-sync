package main

import (
	"fmt"

	"github.com/smelt02/touhou-local-sync/internal/models"
	"github.com/smelt02/touhou-local-sync/pkg/config"
	"github.com/smelt02/touhou-local-sync/pkg/device"
	"github.com/smelt02/touhou-local-sync/pkg/logger"
	"github.com/smelt02/touhou-local-sync/pkg/pathdetect"
	"github.com/smelt02/touhou-local-sync/pkg/sync"
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
	case "CONFLICT":
		fmt.Printf("⚠ %s: Conflict detected (%s)\n", title, comparison.Reason)
	}

	return nil
}
