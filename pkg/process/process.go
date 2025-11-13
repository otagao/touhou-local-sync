// Package process handles game process detection and file lock checking.
// Windows-specific implementation.
package process

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First = kernel32.NewProc("Process32FirstW")
	procProcess32Next  = kernel32.NewProc("Process32NextW")
)

const (
	TH32CS_SNAPPROCESS         = 0x00000002
	MAX_PATH                   = 260
	ERROR_SHARING_VIOLATION    = syscall.Errno(32)
)

// PROCESSENTRY32 represents a process entry in Windows.
type PROCESSENTRY32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [MAX_PATH]uint16
}

// IsProcessRunning checks if a process with the given name is currently running.
// processName should include the .exe extension (e.g., "th08.exe").
func IsProcessRunning(processName string) (bool, error) {
	processName = strings.ToLower(processName)

	// Create snapshot of all processes
	handle, _, err := procCreateToolhelp.Call(TH32CS_SNAPPROCESS, 0)
	if handle == 0 || handle == uintptr(syscall.InvalidHandle) {
		return false, fmt.Errorf("failed to create process snapshot: %w", err)
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	// Iterate through processes
	var entry PROCESSENTRY32
	entry.dwSize = uint32(unsafe.Sizeof(entry))

	// Get first process
	ret, _, err := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return false, fmt.Errorf("failed to get first process: %w", err)
	}

	// Check first process
	exeName := strings.ToLower(syscall.UTF16ToString(entry.szExeFile[:]))
	if exeName == processName {
		return true, nil
	}

	// Iterate through remaining processes
	for {
		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}

		exeName := strings.ToLower(syscall.UTF16ToString(entry.szExeFile[:]))
		if exeName == processName {
			return true, nil
		}
	}

	return false, nil
}

// IsFileLocked checks if a file is currently locked by another process.
// This attempts to open the file with exclusive access to detect locks.
func IsFileLocked(filePath string) (bool, error) {
	// Check if file exists first
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// Try to open with exclusive access
	// Windows: Use CreateFile with no sharing flags
	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to convert path: %w", err)
	}

	handle, err := syscall.CreateFile(
		pathPtr,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0, // dwShareMode = 0 means exclusive access
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if err != nil {
		// If we get a sharing violation, the file is locked
		if err == ERROR_SHARING_VIOLATION {
			return true, nil
		}
		// Other errors might indicate permission issues
		return false, fmt.Errorf("failed to open file for lock check: %w", err)
	}

	// Successfully opened, file is not locked
	syscall.CloseHandle(handle)
	return false, nil
}

// GetGameProcessName returns the expected process name for a given title.
// For example, "th08" -> "th08.exe"
func GetGameProcessName(title string) string {
	return title + ".exe"
}

// CanSafelyWrite checks if it's safe to write to a file.
// Returns true if the file is not locked and the game is not running.
func CanSafelyWrite(filePath string, title string) (safe bool, reason string, err error) {
	// Check if game process is running
	processName := GetGameProcessName(title)
	running, err := IsProcessRunning(processName)
	if err != nil {
		return false, "", fmt.Errorf("failed to check process: %w", err)
	}
	if running {
		return false, fmt.Sprintf("process_running: %s", processName), nil
	}

	// Check if file is locked
	locked, err := IsFileLocked(filePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to check file lock: %w", err)
	}
	if locked {
		return false, "file_locked", nil
	}

	return true, "", nil
}
