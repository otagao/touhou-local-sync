package main

import (
	"fmt"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/config"
	"github.com/otagao/touhou-local-sync/pkg/device"
	"github.com/otagao/touhou-local-sync/pkg/pathdetect"
	"github.com/spf13/cobra"
)

var (
	detectGameDir string
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "半自動認識 + 対話登録",
	Long: `セーブデータを半自動認識し、対話的に登録します。

前提条件:
  - Windows 10/11の一般的なファイル構造を想定
  - ゲーム本体の実行ファイルは単一の親フォルダ配下にまとめて配置されていること

検出ステップ:
  1. 既知パターンでセーブデータを探索
  2. 見つかった候補を一覧表示
  3. ユーザーが登録するものを選択

未検出タイトルの手動登録:
  検出されなかったタイトルを対話的に追加できます。`,
	RunE: runDetect,
}

func init() {
	detectCmd.Flags().StringVarP(&detectGameDir, "gamedir", "g", "", "ゲームディレクトリのパス（省略可）")
}

func runDetect(cmd *cobra.Command, args []string) error {
	fmt.Println("=== thlocalsync detect ===")
	fmt.Println()

	// Get device ID
	deviceID, macHash, hostname, err := device.GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	fmt.Printf("Device ID: %s\n", deviceID)
	fmt.Printf("Hostname: %s\n", hostname)
	fmt.Println()

	// Load existing configurations
	devicesConfig, err := config.LoadDevices()
	if err != nil {
		return fmt.Errorf("failed to load devices config: %w", err)
	}

	pathsConfig, err := config.LoadPaths()
	if err != nil {
		return fmt.Errorf("failed to load paths config: %w", err)
	}

	// Update device in config
	updateDeviceConfig(devicesConfig, deviceID, hostname, macHash)

	// Detect save files
	fmt.Println("Searching for save files...")
	detectResult, err := pathdetect.DetectSaveFiles(detectGameDir)
	if err != nil {
		return fmt.Errorf("failed to detect save files: %w", err)
	}

	// Display candidates
	pathdetect.DisplayCandidates(detectResult.Candidates)

	// Prompt for selection
	if len(detectResult.Candidates) > 0 {
		indices, err := pathdetect.PromptCandidateSelection(len(detectResult.Candidates))
		if err != nil {
			return fmt.Errorf("failed to read selection: %w", err)
		}

		// Add selected candidates to config
		registered := 0
		for _, index := range indices {
			if index >= 0 && index < len(detectResult.Candidates) {
				candidate := detectResult.Candidates[index]
				pathdetect.AddCandidateToConfig(candidate, deviceID, pathsConfig)
				registered++
				fmt.Printf("Registered: %s -> %s\n", candidate.Title, candidate.Path)
			}
		}

		if registered > 0 {
			fmt.Printf("\nRegistered %d path(s)\n", registered)
		}
	}

	// Handle not found titles
	if len(detectResult.NotFound) > 0 {
		fmt.Println("\n=== Manual Registration ===")
		fmt.Printf("%d title(s) not found automatically.\n\n", len(detectResult.NotFound))

		for _, title := range detectResult.NotFound {
			path, err := pathdetect.PromptManualPath(title)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if path != "" {
				// Add to config
				candidate := models.DetectCandidate{
					Title: title.Code,
					Path:  path,
				}
				pathdetect.AddCandidateToConfig(candidate, deviceID, pathsConfig)
				fmt.Printf("Registered: %s -> %s\n", title.Code, path)
			}
		}
	}

	// Save configurations
	if err := config.SaveDevices(devicesConfig); err != nil {
		return fmt.Errorf("failed to save devices config: %w", err)
	}

	if err := config.SavePaths(pathsConfig); err != nil {
		return fmt.Errorf("failed to save paths config: %w", err)
	}

	fmt.Println("\n✓ Configuration saved")
	return nil
}

// updateDeviceConfig updates or adds a device to the device configuration.
func updateDeviceConfig(config *models.DeviceConfig, deviceID, hostname, macHash string) {
	// Check if device already exists
	found := false
	for i := range config.Devices {
		if config.Devices[i].ID == deviceID {
			// Update existing device
			config.Devices[i].Hostname = hostname
			config.Devices[i].MACHash = macHash
			config.Devices[i].LastSeen = getCurrentTime()
			found = true
			break
		}
	}

	// Add new device if not found
	if !found {
		newDevice := models.Device{
			ID:       deviceID,
			Hostname: hostname,
			MACHash:  macHash,
			LastSeen: getCurrentTime(),
		}
		config.Devices = append(config.Devices, newDevice)
	}
}
