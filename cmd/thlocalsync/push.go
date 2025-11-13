package main

import (
	"fmt"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/config"
	"github.com/otagao/touhou-local-sync/pkg/device"
	"github.com/otagao/touhou-local-sync/pkg/logger"
	"github.com/otagao/touhou-local-sync/pkg/pathdetect"
	"github.com/otagao/touhou-local-sync/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	pushForce bool
)

var pushCmd = &cobra.Command{
	Use:   "push [title|all]",
	Short: "ポータブルストレージ → ローカル（配布）",
	Long: `ポータブルストレージの正本をローカルへ配布します。

ポータブルストレージがローカルより新しい/大きい場合に上書きします。
ゲーム実行中やファイルロック中は書き込みを禁止します。
上書き前にローカル側のファイルはバックアップされます。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPush,
}

func init() {
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "強制的に上書き（警告を無視）")
}

func runPush(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("=== thlocalsync push ===\n")
	fmt.Printf("Device: %s (%s)\n", deviceID, hostname)
	if pushForce {
		fmt.Println("⚠ Force mode enabled")
	}
	fmt.Println()

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

	// Get titles to push
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

	// Push each title
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, title := range titles {
		err := pushTitle(title, deviceID, pathsConfig, log, pushForce)
		if err != nil {
			fmt.Printf("✗ %s: %v\n", title, err)
			errorCount++
			// Log error
			log.Error("push_error", map[string]interface{}{
				"title":  title,
				"device": deviceID,
				"error":  err.Error(),
			})
		} else {
			successCount++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Success: %d, Skipped: %d, Errors: %d\n", successCount, skipCount, errorCount)

	return nil
}

func pushTitle(title, deviceID string, pathsConfig *models.PathsConfig, log *logger.Logger, force bool) error {
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

	// Push file
	comparison, err := sync.PushFile(title, vaultPath, localPath, force)
	if err != nil {
		return err
	}

	// Report result
	switch comparison.Recommendation {
	case "PUSH":
		fmt.Printf("✓ %s: Pushed to local (%s)\n", title, comparison.Reason)
		// Log operation
		log.Info("push", map[string]interface{}{
			"title":  title,
			"device": deviceID,
			"action": "update",
			"from":   "usb",
			"to":     "local",
			"reason": comparison.Reason,
		})
	case "SKIP":
		fmt.Printf("- %s: Skipped (%s)\n", title, comparison.Reason)
	case "PULL":
		fmt.Printf("- %s: Local is newer, skipped (%s)\n", title, comparison.Reason)
	case "CONFLICT":
		fmt.Printf("⚠ %s: Conflict detected (%s)\n", title, comparison.Reason)
	}

	return nil
}
