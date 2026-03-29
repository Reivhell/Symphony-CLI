package remote

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

type FetchMeta struct {
	Source       string
	ResolvedType string // "local", "git", "http"
	Version      string // branch/tag
	Commit       string
	CachedAt     time.Time
}

// Fetch memvalidasi / mengambil target dari cache atau remote origin-nya
func Fetch(source, cacheDir string) (string, FetchMeta, error) {
	// 1. Cek apakah itu sebuah Path Lokal Absolut/Relatif
	if isLocalPath(source) {
		abs, err := filepath.Abs(source)
		if err != nil {
			return "", FetchMeta{}, fmt.Errorf("local path tidak valid: %w", err)
		}
		if _, err := os.Stat(filepath.Join(abs, "template.yaml")); err != nil {
			return "", FetchMeta{}, fmt.Errorf("direktori lokal bukan sebuah symphony template: %s", abs)
		}
		return abs, FetchMeta{Source: abs, ResolvedType: "local"}, nil
	}

	// 2. Jika Valid dari Cache
	if IsCached(source, cacheDir) {
		return CachedPath(source, cacheDir), FetchMeta{Source: source, ResolvedType: "cache_hit"}, nil
	}

	lockPath := filepath.Join(cacheDir, CacheKey(source)+".lock")
	fl := flock.New(lockPath)
	if err := fl.Lock(); err != nil {
		return "", FetchMeta{}, fmt.Errorf("cache lock: %w", err)
	}
	defer func() { _ = fl.Unlock() }()

	if IsCached(source, cacheDir) {
		return CachedPath(source, cacheDir), FetchMeta{Source: source, ResolvedType: "cache_hit"}, nil
	}

	// 3. Tarik ke temp dir
	tmpDir, err := os.MkdirTemp("", "symphony-fetch-*")
	if err != nil {
		return "", FetchMeta{}, err
	}
	// Pastikan kita bisa menghapus tmpDir jika terjadi error (hindari sampah disk)
	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()

	var meta FetchMeta
	meta.Source = source
	meta.CachedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 4. Deteksi Remote Protocol
	if isHTTPArchive(source) {
		meta.ResolvedType = "http"
		err = fetchHTTPZip(ctx, source, tmpDir)
	} else if isGitSource(source) {
		meta.ResolvedType = "git"
		// Git mem-pull langsung ke folder tmp, tapi format clone-nya butuh git installed
		err = fetchGit(ctx, source, tmpDir, &meta)
	} else {
		err = fmt.Errorf("format source tidak dikenal: %s", source)
	}

	if err != nil {
		return "", FetchMeta{}, err
	}

	// 5. Verifikasi Template Schema
	if _, errStat := os.Stat(filepath.Join(tmpDir, "template.yaml")); errStat != nil {
		err = fmt.Errorf("bukan symphony template valid (tidak ada template.yaml)")
		return "", FetchMeta{}, err
	}

	// 6. Simpan ke Cache
	finalCacheDir := CachedPath(source, cacheDir)
	_ = os.RemoveAll(finalCacheDir) // hapus jika ada residu
	
	if renameErr := os.Rename(tmpDir, finalCacheDir); renameErr != nil {
		// Di Windows tmp ke cache mungkin ada di drive berbeda
		err = fmt.Errorf("gagal memindahkan ke cache: %v", renameErr)
		return "", FetchMeta{}, err
	}

	// 7. Hitung size & write meta
	size, _ := dirSize(finalCacheDir)
	isTagged := meta.Version != "" && !strings.Contains(meta.Version, "main") && !strings.Contains(meta.Version, "master")
	if meta.Version != "" && strings.HasPrefix(meta.Version, "v") {
		isTagged = true
	}
	
	entry := CacheEntry{
		Source:    source,
		LocalPath: finalCacheDir,
		CachedAt:  meta.CachedAt,
		SizeBytes: size,
		IsTagged:  isTagged,
	}
	WriteMeta(source, cacheDir, entry)

	return finalCacheDir, meta, nil
}

func isLocalPath(s string) bool {
	return strings.HasPrefix(s, ".") || strings.HasPrefix(s, "/") || strings.Contains(s, ":\\")
}

func isHTTPArchive(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return strings.HasSuffix(u.Path, ".zip")
}

func isGitSource(s string) bool {
	// Bisa berupa github.com/user/repo, githab.com/user/repo@v1, dsb
	if strings.Contains(s, "github.com") || strings.Contains(s, "gitlab.com") || strings.Contains(s, "bitbucket.org") {
		return true
	}
	return strings.HasPrefix(s, "git@") || strings.HasPrefix(s, "http")
}

func fetchHTTPZip(ctx context.Context, uri, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return err
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return fmt.Errorf("DNS/network error (check internet and DNS): %w", err)
		}
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return fmt.Errorf("network error (check your connection): %w", err)
		}
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("server returned 403 (rate limit or forbidden); set GITHUB_TOKEN or retry later")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP status %d", resp.StatusCode)
	}

	// Untuk zip ekstraksi di go kita butuh random access byte raider, simpan sementara:
	tmpZip, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpZip.Name())

	if _, err := io.Copy(tmpZip, resp.Body); err != nil {
		return fmt.Errorf("download interrupted (partial file discarded): %w", err)
	}
	tmpZip.Close()

	r, err := zip.OpenReader(tmpZip.Name())
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("terjadi direktori traversal file ilegal: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, errCopy := io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if errCopy != nil {
			return errCopy
		}
	}

	return nil
}

func fetchGit(ctx context.Context, source, dest string, meta *FetchMeta) error {
	parts := strings.SplitN(source, "@", 2)
	repoURL := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}

	meta.Version = version

	if !strings.HasPrefix(repoURL, "http") && !strings.HasPrefix(repoURL, "git@") {
		// Asumsi prefix implicit https:// (missal github.com/symphony/cli)
		repoURL = "https://" + repoURL
	}

	var args []string
	if version != "" {
		args = []string{"clone", "--depth=1", "--branch", version, repoURL, dest}
	} else {
		args = []string{"clone", "--depth=1", repoURL, dest}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	
	var errBuffer bytes.Buffer
	cmd.Stderr = &errBuffer

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gagal eksekusi git: %w. %s", err, errBuffer.String())
	}
	
	// resolve commit sha (optional best-effort)
	cmd2 := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd2.Dir = dest
	if out, err := cmd2.Output(); err == nil {
		meta.Commit = strings.TrimSpace(string(out))
	}
	
	// Hapus folder .git
	os.RemoveAll(filepath.Join(dest, ".git"))

	return nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
