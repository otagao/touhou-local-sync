package main

import (
	"fmt"

	"github.com/smelt02/touhou-local-sync/pkg/backup"
	"github.com/smelt02/touhou-local-sync/pkg/pathdetect"
	"github.com/smelt02/touhou-local-sync/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	backupList    bool
	backupRestore string
)

var backupCmd = &cobra.Command{
	Use:   "backup [title]",
	Short: "履歴表示/復元",
	Long: `セーブデータのバックアップ履歴を表示または復元します。

使用例:
  thlocalsync backup th08 --list          履歴一覧を表示
  thlocalsync backup th08 --restore <name> 指定バックアップを復元`,
	Args: cobra.ExactArgs(1),
	RunE: runBackup,
}

func init() {
	backupCmd.Flags().BoolVarP(&backupList, "list", "l", false, "バックアップ履歴を一覧表示")
	backupCmd.Flags().StringVarP(&backupRestore, "restore", "r", "", "指定バックアップを復元")
}

func runBackup(cmd *cobra.Command, args []string) error {
	title := args[0]

	// Validate title code
	if !pathdetect.IsValidTitleCode(title) {
		return fmt.Errorf("invalid title code: %s", title)
	}

	fmt.Printf("=== thlocalsync backup: %s ===\n\n", title)

	// Determine vault file name
	titleInfo := pathdetect.GetTitleByCode(title)
	var fileName string
	if titleInfo != nil {
		fileName = titleInfo.FileName
	} else {
		fileName = "score.dat"
	}

	// Get vault path for restoration target
	vaultPath, err := sync.GetVaultFilePath(title, fileName)
	if err != nil {
		return fmt.Errorf("failed to get vault path: %w", err)
	}

	// List backups
	if backupList || backupRestore == "" {
		details, err := backup.GetBackupDetails(title)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(details) == 0 {
			fmt.Println("No backups found.")
			return nil
		}

		fmt.Printf("Found %d backup(s):\n\n", len(details))
		for i, detail := range details {
			fmt.Printf("[%d] %s\n", i+1, detail.Name)
			if !detail.Timestamp.IsZero() {
				fmt.Printf("    Time: %s\n", detail.Timestamp.Format("2006-01-02 15:04:05 MST"))
			}
			if detail.Size > 0 {
				fmt.Printf("    Size: %d bytes\n", detail.Size)
			}
			if detail.Error != nil {
				fmt.Printf("    Error: %v\n", detail.Error)
			}
			fmt.Println()
		}

		return nil
	}

	// Restore backup
	if backupRestore != "" {
		fmt.Printf("Restoring backup: %s\n", backupRestore)

		err := backup.RestoreBackup(title, backupRestore, vaultPath)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		fmt.Printf("✓ Successfully restored %s to vault\n", backupRestore)
		fmt.Printf("  Target: %s\n", vaultPath)

		return nil
	}

	return nil
}
