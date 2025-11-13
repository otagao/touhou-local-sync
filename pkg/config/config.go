// Package config handles JSON configuration file I/O.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/utils"
)

const (
	// ConfigDir is the relative path to the config directory from the executable
	ConfigDir = "data"

	// DevicesFile is the filename for device configuration
	DevicesFile = "devices.json"

	// PathsFile is the filename for path configuration
	PathsFile = "paths.json"

	// RulesFile is the filename for sync rules
	RulesFile = "rules.json"
)

// GetConfigDir returns the absolute path to the config directory.
// It assumes the config directory is relative to the executable location.
func GetConfigDir() (string, error) {
	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get directory containing executable
	exeDir := filepath.Dir(exePath)

	// Config directory is <exe_dir>/data
	configDir := filepath.Join(exeDir, ConfigDir)

	return configDir, nil
}

// LoadDevices loads the devices.json configuration.
// If the file doesn't exist, returns an empty config.
func LoadDevices() (*models.DeviceConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(configDir, DevicesFile)

	// If file doesn't exist, return empty config
	exists, _ := utils.FileExists(filePath)
	if !exists {
		return &models.DeviceConfig{Devices: []models.Device{}}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read devices.json: %w", err)
	}

	var config models.DeviceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Backup corrupted file
		backupPath := filePath + ".backup-" + time.Now().Format("20060102-150405")
		_ = utils.AtomicCopy(filePath, backupPath)
		return nil, fmt.Errorf("failed to parse devices.json (backed up to %s): %w", backupPath, err)
	}

	return &config, nil
}

// SaveDevices saves the devices.json configuration atomically.
func SaveDevices(config *models.DeviceConfig) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := utils.EnsureDir(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, DevicesFile)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal devices config: %w", err)
	}

	// Write to temp file first
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// LoadPaths loads the paths.json configuration.
// If the file doesn't exist, returns an empty config.
func LoadPaths() (*models.PathsConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(configDir, PathsFile)

	// If file doesn't exist, return empty config
	exists, _ := utils.FileExists(filePath)
	if !exists {
		return &models.PathsConfig{
			Paths: make(map[string]map[string]models.PathEntry),
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read paths.json: %w", err)
	}

	var config models.PathsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Backup corrupted file
		backupPath := filePath + ".backup-" + time.Now().Format("20060102-150405")
		_ = utils.AtomicCopy(filePath, backupPath)
		return nil, fmt.Errorf("failed to parse paths.json (backed up to %s): %w", backupPath, err)
	}

	// Ensure Paths map is initialized
	if config.Paths == nil {
		config.Paths = make(map[string]map[string]models.PathEntry)
	}

	return &config, nil
}

// SavePaths saves the paths.json configuration atomically.
func SavePaths(config *models.PathsConfig) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := utils.EnsureDir(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, PathsFile)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal paths config: %w", err)
	}

	// Write to temp file first
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// LoadRules loads the rules.json configuration.
// If the file doesn't exist, returns default rules.
func LoadRules() (*models.Rules, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(configDir, RulesFile)

	// If file doesn't exist, return default config
	exists, _ := utils.FileExists(filePath)
	if !exists {
		return &models.Rules{
			Include:      []string{"score.dat", "scoreth*.dat"},
			Exclude:      []string{"*.tmp", "_history/*"},
			HistoryLimit: 20,
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules.json: %w", err)
	}

	var config models.Rules
	if err := json.Unmarshal(data, &config); err != nil {
		// Backup corrupted file
		backupPath := filePath + ".backup-" + time.Now().Format("20060102-150405")
		_ = utils.AtomicCopy(filePath, backupPath)
		return nil, fmt.Errorf("failed to parse rules.json (backed up to %s): %w", backupPath, err)
	}

	return &config, nil
}

// SaveRules saves the rules.json configuration atomically.
func SaveRules(config *models.Rules) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := utils.EnsureDir(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, RulesFile)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules config: %w", err)
	}

	// Write to temp file first
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
