// Package models defines internal data structures used across the application.
package models

import "time"

// Device represents a PC/device that uses this sync tool.
type Device struct {
	ID       string    `json:"id"`        // SHA256(hostname+mac) の先頭12文字
	Hostname string    `json:"hostname"`  // PC名
	MACHash  string    `json:"mac_hash"`  // "sha256:..." 形式
	LastSeen time.Time `json:"last_seen"` // 最終接続時刻
}

// DeviceConfig represents the devices.json structure.
type DeviceConfig struct {
	Devices []Device `json:"devices"`
}

// PathEntry represents a single path configuration for a title on a specific device.
type PathEntry struct {
	Paths     []string `json:"paths"`     // 複数パス候補（環境変数展開前）
	Preferred int      `json:"preferred"` // 優先パスのインデックス
}

// PathsConfig represents the paths.json structure.
// Map: title -> device_id -> PathEntry
type PathsConfig struct {
	Paths map[string]map[string]PathEntry `json:"paths"` // title -> device_id -> PathEntry
}

// Rules represents the rules.json structure.
type Rules struct {
	Include      []string `json:"include"`       // 同期対象パターン
	Exclude      []string `json:"exclude"`       // 除外パターン
	HistoryLimit int      `json:"history_limit"` // 履歴保存上限
}

// FileMetadata contains file information for comparison.
type FileMetadata struct {
	Path     string    // 絶対パス
	Exists   bool      // ファイル存在
	Readable bool      // 読み取り可能
	Size     int64     // サイズ（バイト）
	ModTime  time.Time // 最終更新時刻（UTC）
	Hash     string    // SHA256ハッシュ（フル）
}

// HashShort returns the first 12 characters of the hash for display.
func (fm *FileMetadata) HashShort() string {
	if len(fm.Hash) < 12 {
		return fm.Hash
	}
	return fm.Hash[:12]
}

// ComparisonResult represents the result of comparing two files.
type ComparisonResult struct {
	LocalMeta     *FileMetadata
	RemoteMeta    *FileMetadata
	HashMatch     bool   // ハッシュ一致
	SizeDiff      int64  // サイズ差（Local - Remote）
	TimeDiff      int64  // 時間差（秒、Local - Remote）
	Recommendation string // "PULL", "PUSH", "SKIP", "CONFLICT"
	Reason        string // 判定理由
}

// SyncOperation represents a single sync operation for logging.
type SyncOperation struct {
	OpID      string    `json:"op_id"`      // UUID
	Timestamp time.Time `json:"time"`       // 実行時刻
	Title     string    `json:"title"`      // タイトル（th06等）
	DeviceID  string    `json:"device"`     // デバイスID
	Action    string    `json:"action"`     // "update", "skip", "backup"
	From      string    `json:"from"`       // "local" or "usb"
	To        string    `json:"to"`         // "usb" or "local"
	Reason    string    `json:"reason"`     // 理由
	Success   bool      `json:"success"`    // 成功/失敗
	Error     string    `json:"error,omitempty"` // エラーメッセージ
}

// DetectCandidate represents a detected save file candidate.
type DetectCandidate struct {
	Title    string        // タイトルコード（th06等）
	Path     string        // 絶対パス
	Metadata *FileMetadata // ファイル情報
}
