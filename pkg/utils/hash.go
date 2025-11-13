// Package utils provides utility functions for hashing, atomic file operations, and time handling.
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateFileHash computes the SHA256 hash of a file.
// Returns the hex-encoded hash string, or an error if the file cannot be read.
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

// CalculateStringHash computes the SHA256 hash of a string.
// Returns the hex-encoded hash string.
func CalculateStringHash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
