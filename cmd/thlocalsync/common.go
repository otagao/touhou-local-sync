package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/otagao/touhou-local-sync/internal/models"
)

// getCurrentTime returns the current time in UTC.
func getCurrentTime() time.Time {
	return time.Now().UTC()
}

// promptUserForConflictResolution asks the user to choose between local, remote, or cancel when a conflict is detected.
// Returns: "local", "remote", or "cancel"
func promptUserForConflictResolution(title string, comparison *models.ComparisonResult, operation string) string {
	fmt.Printf("\n⚠ Conflict detected for %s:\n", title)
	fmt.Printf("   %s\n\n", comparison.Reason)

	fmt.Println("File details:")
	fmt.Printf("  Local:  size=%d, mtime=%s, hash=%s\n",
		comparison.LocalMeta.Size,
		comparison.LocalMeta.ModTime.Format("2006-01-02 15:04:05"),
		truncateHash(comparison.LocalMeta.Hash))
	fmt.Printf("  Remote: size=%d, mtime=%s, hash=%s\n",
		comparison.RemoteMeta.Size,
		comparison.RemoteMeta.ModTime.Format("2006-01-02 15:04:05"),
		truncateHash(comparison.RemoteMeta.Hash))

	fmt.Println("\nWhich file should be used?")
	if operation == "pull" {
		fmt.Println("  [l] Use local file (pull to USB)")
		fmt.Println("  [r] Use remote file (keep USB version)")
	} else {
		fmt.Println("  [l] Use local file (keep local version)")
		fmt.Println("  [r] Use remote file (push from USB)")
	}
	fmt.Println("  [c] Cancel this operation")
	fmt.Print("\nYour choice [l/r/c]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "cancel"
	}

	input = strings.ToLower(strings.TrimSpace(input))
	switch input {
	case "l", "local":
		return "local"
	case "r", "remote":
		return "remote"
	case "c", "cancel":
		return "cancel"
	default:
		fmt.Println("Invalid choice, cancelling.")
		return "cancel"
	}
}

// truncateHash returns the first 12 characters of a hash for display.
func truncateHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}
