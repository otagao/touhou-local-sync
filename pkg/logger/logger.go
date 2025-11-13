// Package logger provides structured JSON Lines logging functionality.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otagao/touhou-local-sync/pkg/utils"
)

const (
	// LogDir is the relative path to the log directory from the executable
	LogDir = "logs"
)

// Level represents the log level.
type Level string

const (
	// LevelInfo represents informational messages
	LevelInfo Level = "INFO"
	// LevelWarn represents warning messages
	LevelWarn Level = "WARN"
	// LevelError represents error messages
	LevelError Level = "ERROR"
)

// Entry represents a single log entry.
type Entry struct {
	Level   Level                  `json:"level"`
	Time    time.Time              `json:"time"`
	Message string                 `json:"msg"`
	Fields  map[string]interface{} `json:",inline"`
}

// Logger handles logging operations.
type Logger struct {
	logDir string
}

// New creates a new logger instance.
func New() (*Logger, error) {
	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get directory containing executable
	exeDir := filepath.Dir(exePath)

	// Log directory is <exe_dir>/logs
	logDir := filepath.Join(exeDir, LogDir)

	// Ensure log directory exists
	if err := utils.EnsureDir(logDir); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Logger{logDir: logDir}, nil
}

// getLogFilePath returns the path to the log file for the current date.
func (l *Logger) getLogFilePath() string {
	today := time.Now().Format("2006-01-02")
	return filepath.Join(l.logDir, today+".log")
}

// log writes a log entry to the appropriate log file.
func (l *Logger) log(level Level, message string, fields map[string]interface{}) error {
	entry := Entry{
		Level:   level,
		Time:    time.Now().UTC(),
		Message: message,
		Fields:  fields,
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Append newline for JSON Lines format
	data = append(data, '\n')

	// Open log file in append mode
	logFile := l.getLogFilePath()
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Write log entry
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// Info logs an informational message.
func (l *Logger) Info(message string, fields map[string]interface{}) error {
	return l.log(LevelInfo, message, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(message string, fields map[string]interface{}) error {
	return l.log(LevelWarn, message, fields)
}

// Error logs an error message.
func (l *Logger) Error(message string, fields map[string]interface{}) error {
	return l.log(LevelError, message, fields)
}

// LogOperation logs a sync operation using SyncOperation model.
func (l *Logger) LogOperation(level Level, op map[string]interface{}) error {
	return l.log(level, op["msg"].(string), op)
}
