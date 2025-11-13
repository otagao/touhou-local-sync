package main

import (
	"fmt"
	"strings"

	"github.com/smelt02/touhou-local-sync/internal/models"
	"github.com/smelt02/touhou-local-sync/pkg/config"
	"github.com/smelt02/touhou-local-sync/pkg/device"
	"github.com/smelt02/touhou-local-sync/pkg/pathdetect"
	"github.com/smelt02/touhou-local-sync/pkg/sync"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [title|all]",
	Short: "ポータブルストレージとローカルの差分一覧",
	Long: `ポータブルストレージとローカルの差分を一覧表示します。

各ファイルのサイズ、更新時刻、ハッシュを比較し、
推奨アクション（PULL/PUSH/SKIP）を表示します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("=== thlocalsync status ===\n")
	fmt.Printf("Device: %s (%s)\n\n", deviceID, hostname)

	// Load configurations
	pathsConfig, err := config.LoadPaths()
	if err != nil {
		return fmt.Errorf("failed to load paths config: %w", err)
	}

	// Get titles to check
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

	// Print header
	fmt.Printf("%-8s %-35s %-35s %-25s\n",
		"Title", "Local(best)", "USB(main)", "Recommendation")
	fmt.Println(strings.Repeat("-", 110))

	// Check each title
	for _, title := range titles {
		err := printTitleStatus(title, deviceID, pathsConfig)
		if err != nil {
			fmt.Printf("%-8s ERROR: %v\n", title, err)
		}
	}

	return nil
}

func printTitleStatus(title, deviceID string, pathsConfig *models.PathsConfig) error {
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
		// Default to score.dat
		fileName = "score.dat"
	}

	// Get vault path
	vaultPath, err := sync.GetVaultFilePath(title, fileName)
	if err != nil {
		return fmt.Errorf("failed to get vault path: %w", err)
	}

	// Get metadata for both files
	localMeta, err := sync.GetFileMetadata(localPath)
	if err != nil {
		return fmt.Errorf("failed to get local metadata: %w", err)
	}

	vaultMeta, err := sync.GetFileMetadata(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to get vault metadata: %w", err)
	}

	// Compare files
	comparison := sync.CompareFiles(localMeta, vaultMeta)

	// Format local info
	localInfo := formatFileInfo(localMeta)
	vaultInfo := formatFileInfo(vaultMeta)

	// Format recommendation
	recommendation := formatRecommendation(comparison)

	fmt.Printf("%-8s %-35s %-35s %-25s\n",
		title, localInfo, vaultInfo, recommendation)

	return nil
}

func formatFileInfo(meta *models.FileMetadata) string {
	if !meta.Exists {
		return "[NOT EXIST]"
	}
	if !meta.Readable {
		return "[NOT READABLE]"
	}

	return fmt.Sprintf("size=%d m=%s h=%s",
		meta.Size,
		meta.ModTime.Format("06-01-02 15:04"),
		meta.HashShort())
}

func formatRecommendation(comparison *models.ComparisonResult) string {
	switch comparison.Recommendation {
	case "PULL":
		return fmt.Sprintf("→ PULL (%s)", shortenReason(comparison.Reason))
	case "PUSH":
		return fmt.Sprintf("← PUSH (%s)", shortenReason(comparison.Reason))
	case "SKIP":
		return "= SKIP (identical)"
	case "CONFLICT":
		return fmt.Sprintf("⚠ CONFLICT (%s)", shortenReason(comparison.Reason))
	default:
		return comparison.Recommendation
	}
}

func shortenReason(reason string) string {
	// Shorten reason for display
	if len(reason) > 40 {
		return reason[:37] + "..."
	}
	return reason
}
