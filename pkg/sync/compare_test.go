package sync

import (
	"testing"
	"time"

	"github.com/otagao/touhou-local-sync/internal/models"
)

func TestCompareFiles_EvidenceConflict(t *testing.T) {
	baseTime := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		localSize      int64
		localTime      time.Time
		remoteSize     int64
		remoteTime     time.Time
		expectedRec    string
		expectedReason string
	}{
		{
			name:        "Local larger and newer - should PULL",
			localSize:   2000,
			localTime:   baseTime.Add(10 * time.Minute),
			remoteSize:  1000,
			remoteTime:  baseTime,
			expectedRec: "PULL",
		},
		{
			name:        "Remote larger and newer - should PUSH",
			localSize:   1000,
			localTime:   baseTime,
			remoteSize:  2000,
			remoteTime:  baseTime.Add(10 * time.Minute),
			expectedRec: "PUSH",
		},
		{
			name:        "Local larger but remote newer - CONFLICT",
			localSize:   2000,
			localTime:   baseTime,
			remoteSize:  1000,
			remoteTime:  baseTime.Add(10 * time.Minute),
			expectedRec: "CONFLICT",
		},
		{
			name:        "Remote larger but local newer - CONFLICT",
			localSize:   1000,
			localTime:   baseTime.Add(10 * time.Minute),
			remoteSize:  2000,
			remoteTime:  baseTime,
			expectedRec: "CONFLICT",
		},
		{
			name:        "Same size, local newer - should PULL",
			localSize:   1500,
			localTime:   baseTime.Add(10 * time.Minute),
			remoteSize:  1500,
			remoteTime:  baseTime,
			expectedRec: "PULL",
		},
		{
			name:        "Same size, remote newer - should PUSH",
			localSize:   1500,
			localTime:   baseTime,
			remoteSize:  1500,
			remoteTime:  baseTime.Add(10 * time.Minute),
			expectedRec: "PUSH",
		},
		{
			name:        "Local larger, time within drift - should PULL",
			localSize:   2000,
			localTime:   baseTime.Add(2 * time.Second),
			remoteSize:  1000,
			remoteTime:  baseTime,
			expectedRec: "PULL",
		},
		{
			name:        "Remote larger, time within drift - should PUSH",
			localSize:   1000,
			localTime:   baseTime,
			remoteSize:  2000,
			remoteTime:  baseTime.Add(2 * time.Second),
			expectedRec: "PUSH",
		},
		{
			name:        "Same size, time within drift - should SKIP",
			localSize:   1500,
			localTime:   baseTime.Add(2 * time.Second),
			remoteSize:  1500,
			remoteTime:  baseTime,
			expectedRec: "SKIP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := &models.FileMetadata{
				Path:     "/local/test.dat",
				Exists:   true,
				Readable: true,
				Size:     tt.localSize,
				ModTime:  tt.localTime,
				Hash:     "local_hash_123",
			}

			remote := &models.FileMetadata{
				Path:     "/remote/test.dat",
				Exists:   true,
				Readable: true,
				Size:     tt.remoteSize,
				ModTime:  tt.remoteTime,
				Hash:     "remote_hash_456",
			}

			result := CompareFiles(local, remote)

			if result.Recommendation != tt.expectedRec {
				t.Errorf("Expected recommendation %s, got %s. Reason: %s",
					tt.expectedRec, result.Recommendation, result.Reason)
			}

			// Log the reason for debugging
			t.Logf("Recommendation: %s, Reason: %s", result.Recommendation, result.Reason)
		})
	}
}

func TestCompareFiles_HashMatch(t *testing.T) {
	baseTime := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)

	local := &models.FileMetadata{
		Path:     "/local/test.dat",
		Exists:   true,
		Readable: true,
		Size:     1000,
		ModTime:  baseTime,
		Hash:     "same_hash_123",
	}

	remote := &models.FileMetadata{
		Path:     "/remote/test.dat",
		Exists:   true,
		Readable: true,
		Size:     1000,
		ModTime:  baseTime,
		Hash:     "same_hash_123",
	}

	result := CompareFiles(local, remote)

	if result.Recommendation != "SKIP" {
		t.Errorf("Expected SKIP for identical files, got %s", result.Recommendation)
	}

	if !result.HashMatch {
		t.Error("Expected HashMatch to be true")
	}
}

func TestCompareFiles_SuspiciouslySizeRatio(t *testing.T) {
	baseTime := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		localSize   int64
		remoteSize  int64
		expectedRec string
	}{
		{
			name:        "Local 3x larger - CONFLICT",
			localSize:   3000,
			remoteSize:  1000,
			expectedRec: "CONFLICT",
		},
		{
			name:        "Remote 3x larger - CONFLICT",
			localSize:   1000,
			remoteSize:  3000,
			expectedRec: "CONFLICT",
		},
		{
			name:        "Local 1.5x larger - should PULL",
			localSize:   1500,
			remoteSize:  1000,
			expectedRec: "PULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := &models.FileMetadata{
				Path:     "/local/test.dat",
				Exists:   true,
				Readable: true,
				Size:     tt.localSize,
				ModTime:  baseTime,
				Hash:     "local_hash",
			}

			remote := &models.FileMetadata{
				Path:     "/remote/test.dat",
				Exists:   true,
				Readable: true,
				Size:     tt.remoteSize,
				ModTime:  baseTime,
				Hash:     "remote_hash",
			}

			result := CompareFiles(local, remote)

			if result.Recommendation != tt.expectedRec {
				t.Errorf("Expected %s, got %s. Reason: %s",
					tt.expectedRec, result.Recommendation, result.Reason)
			}
		})
	}
}
