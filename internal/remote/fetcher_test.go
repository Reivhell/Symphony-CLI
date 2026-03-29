package remote

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockZip mengembalikan byte mock zip archive dasar
func createMockZip() []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("template.yaml")
	f.Write([]byte(`schema_version: "2"\nname: "test"\nversion: "1.0"\nactions: []`))
	
	f2, _ := w.Create("main.go")
	f2.Write([]byte("package main\nfunc main() {}"))

	w.Close()
	return buf.Bytes()
}

func TestFetch_LocalValid(t *testing.T) {
	// Setup lokal source folder
	sourceDir := t.TempDir()
	yamlPath := filepath.Join(sourceDir, "template.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("name: test"), 0644))

	cacheDir := t.TempDir()

	localPath, meta, err := Fetch(sourceDir, cacheDir)
	assert.NoError(t, err)
	assert.Equal(t, "local", meta.ResolvedType)
	assert.NotEmpty(t, localPath)
	assert.Equal(t, sourceDir, localPath) // Jika lokal path, ia tak cache, ia mem-bypass langsung ke sourceDir ini.
}

func TestFetch_LocalInvalid(t *testing.T) {
	sourceDir := t.TempDir()
	cacheDir := t.TempDir()

	// Tanpa template.yaml!
	_, _, err := Fetch(sourceDir, cacheDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "direktori lokal bukan sebuah symphony template")
}

func TestFetch_HTTPArchive(t *testing.T) {
	mockZip := createMockZip()

	// Setup Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(mockZip)
	}))
	defer ts.Close()

	cacheDir := t.TempDir()
	sourceURL := ts.URL + "/template.zip"

	localPath, meta, err := Fetch(sourceURL, cacheDir)
	require.NoError(t, err)
	assert.Equal(t, "http", meta.ResolvedType)
	assert.NotEmpty(t, localPath)

	// Pastikan file tertulis
	_, err = os.Stat(filepath.Join(localPath, "template.yaml"))
	assert.NoError(t, err)
	
	// Validasi via Cached() check manual
	list, _ := List(cacheDir)
	assert.Len(t, list, 1)
	assert.Equal(t, sourceURL, list[0].Source)
}

func TestFetch_HTTPMissingTemplate(t *testing.T) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	w.Create("hello.txt") // zip tanpa template.yaml
	w.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buf.Bytes())
	}))
	defer ts.Close()

	cacheDir := t.TempDir()
	sourceURL := ts.URL + "/repo.zip"

	_, _, err := Fetch(sourceURL, cacheDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ada template.yaml")
}

func TestIsLocalPath(t *testing.T) {
	assert.True(t, isLocalPath("./template"))
	assert.True(t, isLocalPath("../parent/dir"))
	assert.True(t, isLocalPath("/usr/local/templates/react"))
	assert.True(t, isLocalPath("C:\\Users\\admin\\tpl"))
	
	assert.False(t, isLocalPath("github.com/user/repo"))
	assert.False(t, isLocalPath("https://foo.com/file.zip"))
}
