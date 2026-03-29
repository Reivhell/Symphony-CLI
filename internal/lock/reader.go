package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Read membaca dan mem-parse file symphony.lock dari direktori yang diberikan.
func Read(dir string) (*LockFile, error) {
	lockPath := filepath.Join(dir, "symphony.lock")
	
	bytes, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("symphony.lock tidak ditemukan di %s: %w", dir, err)
	}

	var lf LockFile
	if err := json.Unmarshal(bytes, &lf); err != nil {
		return nil, fmt.Errorf("symphony.lock corrupt atau tidak valid: %w", err)
	}

	return &lf, nil
}
