package sync

import (
	"fmt"

	"github.com/otagao/touhou-local-sync/internal/models"
	"github.com/otagao/touhou-local-sync/pkg/utils"
)

const (
	// MaxSizeRatio is the maximum acceptable size ratio (new/old) before flagging as suspicious
	MaxSizeRatio = 2.0
)

// CompareFiles performs a three-point comparison (hash, size, mtime) between two files.
// Returns a ComparisonResult with recommendation and reason.
//
// Comparison logic (as per spec §9.2):
// 1. If hash matches → files are identical, SKIP
// 2. If hash differs:
//    a. If size differs → larger file is preferred (with suspicious check)
//    b. If size same but mtime differs → newer mtime is preferred (with drift tolerance)
// 3. Final decision can be overridden by user interaction
func CompareFiles(local, remote *models.FileMetadata) *models.ComparisonResult {
	result := &models.ComparisonResult{
		LocalMeta:  local,
		RemoteMeta: remote,
	}

	// Handle cases where one or both files don't exist
	if !local.Exists && !remote.Exists {
		result.Recommendation = "SKIP"
		result.Reason = "both files do not exist"
		return result
	}

	if !local.Exists {
		result.Recommendation = "PUSH"
		result.Reason = "local file does not exist"
		return result
	}

	if !remote.Exists {
		result.Recommendation = "PULL"
		result.Reason = "remote file does not exist"
		return result
	}

	// Handle readability issues
	if !local.Readable {
		result.Recommendation = "SKIP"
		result.Reason = "local file not readable"
		return result
	}

	if !remote.Readable {
		result.Recommendation = "SKIP"
		result.Reason = "remote file not readable"
		return result
	}

	// Calculate differences
	result.SizeDiff = local.Size - remote.Size
	result.TimeDiff = utils.TimeDiffSeconds(local.ModTime, remote.ModTime)

	// 1. Check hash match
	if local.Hash == remote.Hash {
		result.HashMatch = true
		result.Recommendation = "SKIP"
		result.Reason = "files are identical (hash match)"
		return result
	}

	result.HashMatch = false

	// 2. Hash differs - analyze both size and mtime as equal evidence

	// Determine size preference
	var sizePreference string // "local", "remote", or "equal"
	var sizeRatio float64

	if result.SizeDiff > 0 {
		// Local is larger
		sizePreference = "local"
		if remote.Size > 0 {
			sizeRatio = float64(local.Size) / float64(remote.Size)
		} else {
			sizeRatio = 999.0 // Remote is empty
		}

		if sizeRatio > MaxSizeRatio {
			result.Recommendation = "CONFLICT"
			result.Reason = fmt.Sprintf("local file suspiciously large (%.1fx larger, local=%d remote=%d)", sizeRatio, local.Size, remote.Size)
			return result
		}
	} else if result.SizeDiff < 0 {
		// Remote is larger
		sizePreference = "remote"
		if local.Size > 0 {
			sizeRatio = float64(remote.Size) / float64(local.Size)
		} else {
			sizeRatio = 999.0 // Local is empty
		}

		if sizeRatio > MaxSizeRatio {
			result.Recommendation = "CONFLICT"
			result.Reason = fmt.Sprintf("remote file suspiciously large (%.1fx larger, remote=%d local=%d)", sizeRatio, remote.Size, local.Size)
			return result
		}
	} else {
		// Sizes are equal
		sizePreference = "equal"
	}

	// Determine time preference
	var timePreference string // "local", "remote", or "equal"

	if utils.TimeWithinDrift(local.ModTime, remote.ModTime) {
		timePreference = "equal"
	} else if utils.IsNewerThan(local.ModTime, remote.ModTime) {
		timePreference = "local"
	} else {
		timePreference = "remote"
	}

	// Combine size and time evidence
	// If both agree or one is equal, make a clear recommendation
	// If they conflict, flag as CONFLICT for user confirmation

	if sizePreference == "equal" && timePreference == "equal" {
		// Both equal - files are essentially the same
		result.Recommendation = "SKIP"
		result.Reason = fmt.Sprintf("files appear identical (size=%d, mtime within %ds drift)", local.Size, utils.TimeDriftTolerance)
		return result
	}

	if sizePreference == "local" && timePreference == "local" {
		// Both prefer local - clear PULL
		result.Recommendation = "PULL"
		result.Reason = fmt.Sprintf("local file is both larger and newer (size: local=%d remote=%d, time: local=%s remote=%s)",
			local.Size, remote.Size,
			local.ModTime.Format("2006-01-02 15:04:05"),
			remote.ModTime.Format("2006-01-02 15:04:05"))
		return result
	}

	if sizePreference == "remote" && timePreference == "remote" {
		// Both prefer remote - clear PUSH
		result.Recommendation = "PUSH"
		result.Reason = fmt.Sprintf("remote file is both larger and newer (size: remote=%d local=%d, time: remote=%s local=%s)",
			remote.Size, local.Size,
			remote.ModTime.Format("2006-01-02 15:04:05"),
			local.ModTime.Format("2006-01-02 15:04:05"))
		return result
	}

	if sizePreference == "equal" {
		// Size equal, time differs - use time preference
		if timePreference == "local" {
			result.Recommendation = "PULL"
			result.Reason = fmt.Sprintf("local file is newer (size equal=%d, time: local=%s remote=%s, diff=%ds)",
				local.Size,
				local.ModTime.Format("2006-01-02 15:04:05"),
				remote.ModTime.Format("2006-01-02 15:04:05"),
				result.TimeDiff)
			return result
		} else {
			result.Recommendation = "PUSH"
			result.Reason = fmt.Sprintf("remote file is newer (size equal=%d, time: remote=%s local=%s, diff=%ds)",
				local.Size,
				remote.ModTime.Format("2006-01-02 15:04:05"),
				local.ModTime.Format("2006-01-02 15:04:05"),
				-result.TimeDiff)
			return result
		}
	}

	if timePreference == "equal" {
		// Time equal, size differs - use size preference
		if sizePreference == "local" {
			result.Recommendation = "PULL"
			result.Reason = fmt.Sprintf("local file is larger (size: local=%d remote=%d, time within drift)",
				local.Size, remote.Size)
			return result
		} else {
			result.Recommendation = "PUSH"
			result.Reason = fmt.Sprintf("remote file is larger (size: remote=%d local=%d, time within drift)",
				remote.Size, local.Size)
			return result
		}
	}

	// If we reach here, size and time preferences conflict
	// Example: local is larger but remote is newer, or vice versa
	result.Recommendation = "CONFLICT"
	if sizePreference == "local" && timePreference == "remote" {
		result.Reason = fmt.Sprintf("evidence conflict: local is larger (%d vs %d) but remote is newer (%s vs %s)",
			local.Size, remote.Size,
			remote.ModTime.Format("2006-01-02 15:04:05"),
			local.ModTime.Format("2006-01-02 15:04:05"))
	} else {
		result.Reason = fmt.Sprintf("evidence conflict: remote is larger (%d vs %d) but local is newer (%s vs %s)",
			remote.Size, local.Size,
			local.ModTime.Format("2006-01-02 15:04:05"),
			remote.ModTime.Format("2006-01-02 15:04:05"))
	}
	return result
}
