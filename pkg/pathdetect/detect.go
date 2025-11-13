package pathdetect

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smelt02/touhou-local-sync/internal/models"
	"github.com/smelt02/touhou-local-sync/pkg/sync"
	"github.com/smelt02/touhou-local-sync/pkg/utils"
)

// DetectResult represents the result of detecting save files.
type DetectResult struct {
	Candidates []models.DetectCandidate // Found candidates
	NotFound   []KnownTitle              // Titles not found
}

// DetectSaveFiles searches for save files using known patterns.
// Returns candidates found and titles not found.
func DetectSaveFiles(gameDirOverride string) (*DetectResult, error) {
	result := &DetectResult{
		Candidates: []models.DetectCandidate{},
		NotFound:   []KnownTitle{},
	}

	titles := GetKnownTitles()

	// Ask user for game directory if any title uses it
	var gameDir string
	if gameDirOverride != "" {
		gameDir = gameDirOverride
	} else {
		// Check if any title needs game directory
		needGameDir := false
		for _, title := range titles {
			if title.UseGameDir {
				needGameDir = true
				break
			}
		}

		if needGameDir {
			fmt.Println("Some titles may be installed in a game directory.")
			fmt.Print("Enter game directory path (or press Enter to skip): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			// Remove whitespace and quotes
			gameDir = strings.TrimSpace(input)
			gameDir = strings.Trim(gameDir, "\"")
		}
	}

	// Search for each title
	for _, title := range titles {
		foundPaths := []string{}

		// Search in known patterns
		foundPaths = append(foundPaths, SearchForTitle(title)...)

		// Search in game directory if provided
		if gameDir != "" && title.UseGameDir {
			// Clean the game directory path (remove quotes if present)
			cleanGameDir := strings.Trim(gameDir, "\"")

			// Look for score file in game directory directly
			scorePath := filepath.Join(cleanGameDir, title.FileName)
			if FileExists(scorePath) {
				foundPaths = append(foundPaths, scorePath)
			}

			// Check for title-specific subdirectory (e.g., gameDir/th06/)
			titleDir := filepath.Join(cleanGameDir, title.Code)
			scorePathInTitle := filepath.Join(titleDir, title.FileName)
			if FileExists(scorePathInTitle) {
				foundPaths = append(foundPaths, scorePathInTitle)
			}

			// Also check for game name subdirectory (e.g., gameDir/東方紅魔郷/)
			if title.Name != "" {
				nameDir := filepath.Join(cleanGameDir, title.Name)
				scorePathInName := filepath.Join(nameDir, title.FileName)
				if FileExists(scorePathInName) {
					foundPaths = append(foundPaths, scorePathInName)
				}
			}
		}

		// Create candidates for each found path
		if len(foundPaths) > 0 {
			for _, path := range foundPaths {
				// Get metadata
				meta, err := sync.GetFileMetadata(path)
				if err != nil {
					continue
				}

				candidate := models.DetectCandidate{
					Title:    title.Code,
					Path:     path,
					Metadata: meta,
				}
				result.Candidates = append(result.Candidates, candidate)
			}
		} else {
			result.NotFound = append(result.NotFound, title)
		}
	}

	return result, nil
}

// DisplayCandidates prints detected candidates in a user-friendly format.
func DisplayCandidates(candidates []models.DetectCandidate) {
	if len(candidates) == 0 {
		fmt.Println("No save files detected.")
		return
	}

	fmt.Println("\n[Detect] Found candidates:")
	for i, candidate := range candidates {
		title := GetTitleByCode(candidate.Title)
		titleDisplay := candidate.Title
		if title != nil {
			titleDisplay = FormatTitleDisplay(title.Code, title.Name)
		}

		fmt.Printf("  [%d] %s\n", i+1, titleDisplay)
		fmt.Printf("      Path: %s\n", candidate.Path)

		if candidate.Metadata != nil && candidate.Metadata.Exists {
			fmt.Printf("      Size: %d bytes  ", candidate.Metadata.Size)
			fmt.Printf("ModTime: %s  ", candidate.Metadata.ModTime.Format("2006-01-02 15:04"))
			fmt.Printf("Hash: %s\n", candidate.Metadata.HashShort())
		}
	}
	fmt.Println()
}

// PromptCandidateSelection asks user to select which candidates to register.
// Returns indices of selected candidates.
func PromptCandidateSelection(count int) ([]int, error) {
	fmt.Printf("Select to register: 1-%d (comma-separated), 'a' for all, 's' to skip: ", count)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Handle special cases
	if input == "s" || input == "S" {
		return []int{}, nil
	}

	if input == "a" || input == "A" {
		// Select all
		indices := make([]int, count)
		for i := 0; i < count; i++ {
			indices[i] = i
		}
		return indices, nil
	}

	// Parse comma-separated numbers
	parts := strings.Split(input, ",")
	var indices []int

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var num int
		_, err := fmt.Sscanf(part, "%d", &num)
		if err != nil {
			fmt.Printf("Warning: invalid input '%s', skipping\n", part)
			continue
		}

		// Convert to 0-based index
		index := num - 1
		if index < 0 || index >= count {
			fmt.Printf("Warning: number %d out of range, skipping\n", num)
			continue
		}

		indices = append(indices, index)
	}

	return indices, nil
}

// AddCandidateToConfig adds a candidate to the paths configuration.
func AddCandidateToConfig(candidate models.DetectCandidate, deviceID string, pathsConfig *models.PathsConfig) {
	title := candidate.Title

	// Initialize title map if not exists
	if pathsConfig.Paths == nil {
		pathsConfig.Paths = make(map[string]map[string]models.PathEntry)
	}

	if pathsConfig.Paths[title] == nil {
		pathsConfig.Paths[title] = make(map[string]models.PathEntry)
	}

	// Get or create path entry for this device
	pathEntry, exists := pathsConfig.Paths[title][deviceID]
	if !exists {
		pathEntry = models.PathEntry{
			Paths:     []string{},
			Preferred: 0,
		}
	}

	// Check if path already exists
	pathExists := false
	for _, p := range pathEntry.Paths {
		if utils.ExpandEnvPath(p) == candidate.Path {
			pathExists = true
			break
		}
	}

	if !pathExists {
		pathEntry.Paths = append(pathEntry.Paths, candidate.Path)
		// Set as preferred if it's the first path
		if len(pathEntry.Paths) == 1 {
			pathEntry.Preferred = 0
		}
	}

	pathsConfig.Paths[title][deviceID] = pathEntry
}

// PromptManualPath asks user to manually enter a path for a title.
// Returns the path or empty string if user skips.
func PromptManualPath(title KnownTitle) (string, error) {
	fmt.Printf("\nNo entry for %s (%s). Add manually? [y/N]: ", title.Code, title.Name)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "y" && input != "yes" {
		return "", nil
	}

	fmt.Printf("Enter absolute path for %s %s: ", title.Code, title.FileName)
	pathInput, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read path: %w", err)
	}

	path := strings.TrimSpace(pathInput)
	if path == "" {
		return "", nil
	}

	// Remove surrounding quotes if present
	path = strings.Trim(path, "\"")

	// Expand environment variables
	path = utils.ExpandEnvPath(path)

	// Validate path
	exists, readable := utils.FileExists(path)
	if !exists {
		fmt.Printf("Warning: File does not exist: %s\n", path)
		fmt.Print("Register anyway? [y/N]: ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			return "", nil
		}
	} else if !readable {
		fmt.Printf("Warning: File exists but is not readable: %s\n", path)
		return "", nil
	} else {
		fmt.Println("Validated: OK")
	}

	fmt.Print("Register this path? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm == "n" || confirm == "no" {
		return "", nil
	}

	return path, nil
}
