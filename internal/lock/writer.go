package lock

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LockFile struct {
	SymphonyVersion string           `json:"symphony_version"`
	GeneratedAt     time.Time        `json:"generated_at"`
	Template        TemplateLockInfo `json:"template"`
	Inputs          map[string]any   `json:"inputs"`
	OutputChecksum  string           `json:"output_checksum"`
	// FileChecksums records SHA-256 of generated output files.
	// Key: relative output path (from output dir). Value: "sha256:<hex>".
	FileChecksums map[string]string `json:"file_checksums,omitempty"`
}

type TemplateLockInfo struct {
	Source  string `json:"source"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// Write menulis LockFile ke direktori output sebagai "symphony.lock"
func Write(lockFile *LockFile, outputDir string) error {
	lockPath := filepath.Join(outputDir, "symphony.lock")
	
	bytes, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal me-marshal lock file: %w", err)
	}

	if err := os.WriteFile(lockPath, bytes, 0644); err != nil {
		return fmt.Errorf("gagal menulis symphony.lock: %w", err)
	}

	return nil
}

// fileChecksum computes SHA-256 from content and formats it as "sha256:<hex>".
func fileChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", hash)
}

// FileChecksum is an exported wrapper for computing a stable content checksum.
func FileChecksum(content string) string {
	return fileChecksum(content)
}
