package lock

import (
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
