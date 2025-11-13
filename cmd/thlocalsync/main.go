// Package main is the entry point for thlocalsync CLI application.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "thlocalsync",
	Short: "東方Project セーブデータ同期ツール",
	Long: `thlocalsync - 東方Projectのセーブデータを複数のPC間でポータブルストレージを介して同期するCLIツール

完全オフライン、ポータブルストレージ常駐、単一実行ファイル。
タイトル別の保存パスを半自動認識＋対話的に登録/編集。
mtime・ハッシュ・サイズの三点で新旧/正誤判定。`,
	Version: "0.1.0",
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(backupCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
