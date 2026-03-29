package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CacheEntry struct {
	Source    string    `json:"source"`
	LocalPath string    `json:"local_path"`
	CachedAt  time.Time `json:"cached_at"`
	SizeBytes int64     `json:"size_bytes"`
	IsTagged  bool      `json:"is_tagged"`
}

var DefaultCacheTTL = 24 * time.Hour
var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)

// CacheDir mengembalikan direktori root untuk cache aplikasi.
// Posisinya berada di ~/.symphony/cache
func CacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan user home dir: %w", err)
	}
	dir := filepath.Join(home, ".symphony", "cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori cache: %w", err)
	}
	return dir, nil
}

// CacheKey menghasilkan string yang aman untuk filesystem dari sebuah source path
// Cth: "github.com/user/repo@v1.2" => "github.com_user_repo_v1.2"
func CacheKey(source string) string {
	// Ganti / dan @ dengan underscore, dan buang karakter unik yang lain jika ada
	safe := strings.ReplaceAll(source, "://", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "@", "_")
	return nonAlphanumericRegex.ReplaceAllString(safe, "")
}

// CachedPath mengembalikan path absolut cache dari sebuah source 
func CachedPath(source, cacheDir string) string {
	return filepath.Join(cacheDir, CacheKey(source))
}

// IsCached memeriksa apakah template dengan format source tertentu ada di dalam cache,
// membaca metadata-nya (cache-meta.json) untuk menentukan rules TTL.
func IsCached(source, cacheDir string) bool {
	dir := CachedPath(source, cacheDir)
	if _, err := os.Stat(dir); err != nil {
		return false
	}
	
	metaPath := filepath.Join(dir, "cache-meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}
	
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false
	}

	if entry.IsTagged {
		return true // Version tags hold immutable truths.
	}
	
	return time.Since(entry.CachedAt) <= DefaultCacheTTL
}

// Invalidate menghapus cache template spesifik sekaligus direktori dan isi metasnya
func Invalidate(source, cacheDir string) error {
	dir := CachedPath(source, cacheDir)
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("gagal invalidasi cache: %w", err)
	}
	return nil
}

// WriteMeta menulis metadata JSON terkait cache entry
func WriteMeta(source, cacheDir string, entry CacheEntry) error {
	dir := CachedPath(source, cacheDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	metaPath := filepath.Join(dir, "cache-meta.json")
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

// List mencantumkan apa saja template yang ada di direktori cache.
func List(cacheDir string) ([]CacheEntry, error) {
	var entries []CacheEntry
	
	dirs, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return entries, nil // tidak ada cache yet
		}
		return nil, fmt.Errorf("read cache dir gagal: %w", err)
	}
	
	for _, rawDir := range dirs {
		if !rawDir.IsDir() {
			continue
		}
		
		metaPath := filepath.Join(cacheDir, rawDir.Name(), "cache-meta.json")
		b, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Tidak ada info valid. lewati
		}
		
		var e CacheEntry
		if err := json.Unmarshal(b, &e); err == nil {
			entries = append(entries, e)
		}
	}
	
	return entries, nil
}
