// Package pathdetect handles semi-automatic detection and interactive registration of save file paths.
package pathdetect

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// KnownTitle represents a known Touhou title with its detection patterns.
type KnownTitle struct {
	Code        string   // Title code (e.g., "th06", "th08")
	Name        string   // Display name
	Patterns    []string // Path patterns to search
	UseAppData  bool     // If true, search in %APPDATA%
	UseGameDir  bool     // If true, ask user for game directory
	FileName    string   // Expected filename (e.g., "score.dat")
}

// GetKnownTitles returns a list of known Touhou titles with their detection patterns.
func GetKnownTitles() []KnownTitle {
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	return []KnownTitle{
		// th06-th09: score.dat in game directory, may also be in VirtualStore
		{
			Code:       "th06",
			Name:       "東方紅魔郷",
			UseGameDir: true,
			FileName:   "score.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方紅魔郷\score.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方紅魔郷\score.dat`),
			},
		},
		{
			Code:       "th07",
			Name:       "東方妖々夢",
			UseGameDir: true,
			FileName:   "score.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方妖々夢\score.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方妖々夢\score.dat`),
			},
		},
		{
			Code:       "th08",
			Name:       "東方永夜抄",
			UseGameDir: true,
			FileName:   "score.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方永夜抄\score.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方永夜抄\score.dat`),
			},
		},
		{
			Code:       "th09",
			Name:       "東方花映塚",
			UseGameDir: true,
			FileName:   "score.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方花映塚\score.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方花映塚\score.dat`),
			},
		},
		// th095, th10: scorethXX.dat in game directory, may also be in VirtualStore
		{
			Code:       "th095",
			Name:       "東方文花帖",
			UseGameDir: true,
			FileName:   "scoreth095.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方文花帖\scoreth095.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方文花帖\scoreth095.dat`),
			},
		},
		{
			Code:       "th10",
			Name:       "東方風神録",
			UseGameDir: true,
			FileName:   "scoreth10.dat",
			Patterns: []string{
				filepath.Join(localAppData, `VirtualStore\Program Files\上海アリス幻樂団\東方風神録\scoreth10.dat`),
				filepath.Join(localAppData, `VirtualStore\Program Files (x86)\上海アリス幻樂団\東方風神録\scoreth10.dat`),
			},
		},
		// th11, th12: scorethXX.dat in game directory (no VirtualStore needed)
		{
			Code:       "th11",
			Name:       "東方地霊殿",
			UseGameDir: true,
			FileName:   "scoreth11.dat",
			Patterns:   []string{},
		},
		{
			Code:       "th12",
			Name:       "東方星蓮船",
			UseGameDir: true,
			FileName:   "scoreth12.dat",
			Patterns:   []string{},
		},
		// th125+: scorethXX.dat in AppData/Roaming/ShanghaiAlice
		{
			Code:       "th125",
			Name:       "ダブルスポイラー",
			UseAppData: true,
			FileName:   "scoreth125.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th125\scoreth125.dat`),
			},
		},
		{
			Code:       "th128",
			Name:       "妖精大戦争",
			UseAppData: true,
			FileName:   "scoreth128.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th128\scoreth128.dat`),
			},
		},
		{
			Code:       "th13",
			Name:       "東方神霊廟",
			UseAppData: true,
			FileName:   "scoreth13.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th13\scoreth13.dat`),
			},
		},
		{
			Code:       "th14",
			Name:       "東方輝針城",
			UseAppData: true,
			FileName:   "scoreth14.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th14\scoreth14.dat`),
			},
		},
		{
			Code:       "th143",
			Name:       "弾幕アマノジャク",
			UseAppData: true,
			FileName:   "scoreth143.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th143\scoreth143.dat`),
			},
		},
		{
			Code:       "th15",
			Name:       "東方紺珠伝",
			UseAppData: true,
			FileName:   "scoreth15.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th15\scoreth15.dat`),
			},
		},
		{
			Code:       "th16",
			Name:       "東方天空璋",
			UseAppData: true,
			FileName:   "scoreth16.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th16\scoreth16.dat`),
			},
		},
		{
			Code:       "th165",
			Name:       "秘封ナイトメアダイアリー",
			UseAppData: true,
			FileName:   "scoreth165.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th165\scoreth165.dat`),
			},
		},
		{
			Code:       "th17",
			Name:       "東方鬼形獣",
			UseAppData: true,
			FileName:   "scoreth17.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th17\scoreth17.dat`),
			},
		},
		{
			Code:       "th18",
			Name:       "東方虹龍洞",
			UseAppData: true,
			FileName:   "scoreth18.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th18\scoreth18.dat`),
			},
		},
		{
			Code:       "th185",
			Name:       "バレットフィリア達の闇市場",
			UseAppData: true,
			FileName:   "scoreth185.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th185\scoreth185.dat`),
			},
		},
		{
			Code:       "th19",
			Name:       "東方獣王園",
			UseAppData: true,
			FileName:   "scoreth19.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th19\scoreth19.dat`),
			},
		},
		{
			Code:       "th20",
			Name:       "東方錦上京",
			UseAppData: true,
			FileName:   "scoreth20.dat",
			Patterns: []string{
				filepath.Join(appData, `ShanghaiAlice\th20\scoreth20.dat`),
			},
		},
	}
}

// IsValidTitleCode checks if a string matches the pattern for a Touhou title code.
// Valid formats: th06, th07, ..., th20, th095, th125, th128, th143, th165, th185
func IsValidTitleCode(code string) bool {
	// Match thXX or thXXX format
	matched, _ := regexp.MatchString(`^th\d+$`, code)
	return matched
}

// GetTitleByCode returns the KnownTitle for a given code.
func GetTitleByCode(code string) *KnownTitle {
	titles := GetKnownTitles()
	for i := range titles {
		if titles[i].Code == code {
			return &titles[i]
		}
	}
	return nil
}

// SearchGameDirectoryForScoreDat searches for score.dat files in a game directory.
// Returns a map of title code -> absolute path.
func SearchGameDirectoryForScoreDat(gameDir string) map[string]string {
	results := make(map[string]string)

	// Search for executable files that match th\d+.exe pattern
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return results
	}

	exePattern := regexp.MustCompile(`^(th\d+)\.exe$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := exePattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			titleCode := matches[1]
			title := GetTitleByCode(titleCode)
			if title == nil {
				continue
			}

			// Check if score file exists in the same directory
			scorePath := filepath.Join(gameDir, title.FileName)
			if _, err := os.Stat(scorePath); err == nil {
				results[titleCode] = scorePath
			}

			// Also check in subdirectories with title name
			titleSubDir := filepath.Join(gameDir, titleCode)
			scorePathInSub := filepath.Join(titleSubDir, title.FileName)
			if _, err := os.Stat(scorePathInSub); err == nil {
				results[titleCode] = scorePathInSub
			}
		}
	}

	return results
}

// ExpandPathPatterns expands environment variables in path patterns.
func ExpandPathPatterns(patterns []string) []string {
	expanded := make([]string, len(patterns))
	for i, pattern := range patterns {
		expanded[i] = os.ExpandEnv(pattern)
	}
	return expanded
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SearchForTitle searches for save files for a specific title using known patterns.
// Returns a list of absolute paths where save files were found.
func SearchForTitle(title KnownTitle) []string {
	var found []string

	// Search in known patterns
	for _, pattern := range title.Patterns {
		if FileExists(pattern) {
			found = append(found, pattern)
		}
	}

	return found
}

// GetAllTitleCodes returns a list of all known title codes.
func GetAllTitleCodes() []string {
	titles := GetKnownTitles()
	codes := make([]string, len(titles))
	for i, title := range titles {
		codes[i] = title.Code
	}
	return codes
}

// FormatTitleDisplay returns a formatted string for displaying a title.
func FormatTitleDisplay(code string, name string) string {
	if name != "" {
		return fmt.Sprintf("%s (%s)", code, name)
	}
	return code
}

// SortTitlesByRelease sorts title codes by release order.
// Returns a new sorted slice.
func SortTitlesByRelease(titles []string) []string {
	// Get all known titles to determine order
	knownTitles := GetKnownTitles()
	orderMap := make(map[string]int)
	for i, title := range knownTitles {
		orderMap[title.Code] = i
	}

	// Create a copy to sort
	sorted := make([]string, len(titles))
	copy(sorted, titles)

	// Sort by release order
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			orderI, okI := orderMap[sorted[i]]
			orderJ, okJ := orderMap[sorted[j]]

			// Unknown titles go to the end
			if !okI && okJ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			} else if okI && okJ && orderI > orderJ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
