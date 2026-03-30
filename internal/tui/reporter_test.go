package tui

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Reivhell/symphony/internal/engine"
)

// captureStdout mengalihkan os.Stdout ke buffer sementara untuk test.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	var buf strings.Builder
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

func TestJSONReporter_OnFileCreated(t *testing.T) {
	r := NewJSONReporter(5)
	out := captureStdout(t, func() {
		r.OnFileCreated("cmd/main.go")
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "progress", event["type"])
	assert.Equal(t, "cmd/main.go", event["file"])
	assert.Equal(t, "CREATE", event["action"])
	assert.Equal(t, float64(1), event["current"])
	assert.Equal(t, float64(5), event["total"])
}

func TestJSONReporter_OnFileSkipped(t *testing.T) {
	r := NewJSONReporter(5)
	out := captureStdout(t, func() {
		r.OnFileSkipped("redis/cache.go", "USE_REDIS == false")
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "file_skip", event["type"])
	assert.Equal(t, "redis/cache.go", event["file"])
	assert.Equal(t, "USE_REDIS == false", event["reason"])
}

func TestJSONReporter_OnHookStart(t *testing.T) {
	r := NewJSONReporter(3)
	out := captureStdout(t, func() {
		r.OnHookStart("go mod tidy")
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "hook_start", event["type"])
	assert.Equal(t, "go mod tidy", event["command"])
}

func TestJSONReporter_OnHookOutput(t *testing.T) {
	r := NewJSONReporter(3)
	out := captureStdout(t, func() {
		r.OnHookOutput("go: downloading github.com/lib/pq v1.10.9")
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "hook_output", event["type"])
	assert.Equal(t, "go: downloading github.com/lib/pq v1.10.9", event["line"])
}

func TestJSONReporter_OnComplete(t *testing.T) {
	r := NewJSONReporter(10)
	stats := engine.CompletionStats{
		FilesCreated:  8,
		FilesSkipped:  2,
		FilesModified: 0,
		DurationMs:    1234,
	}
	out := captureStdout(t, func() {
		r.OnComplete(stats)
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "complete", event["type"])
	assert.Equal(t, float64(8), event["files_created"])
	assert.Equal(t, float64(2), event["files_skipped"])
	assert.Equal(t, float64(1234), event["duration_ms"])
}

func TestJSONReporter_OnError(t *testing.T) {
	r := NewJSONReporter(5)
	out := captureStdout(t, func() {
		r.OnError(assert.AnError)
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &event))

	assert.Equal(t, "error", event["type"])
	assert.NotEmpty(t, event["message"])
}

func TestJSONReporter_MultipleEvents(t *testing.T) {
	r := NewJSONReporter(2)
	out := captureStdout(t, func() {
		r.OnFileCreated("file1.go")
		r.OnFileCreated("file2.go")
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Equal(t, 2, len(lines), "harus ada 2 event JSON (NDJSON)")

	for _, line := range lines {
		var event map[string]any
		assert.NoError(t, json.Unmarshal([]byte(line), &event), "setiap baris harus valid JSON")
	}
}

func TestPlainReporter_DoesNotPanic(t *testing.T) {
	r := &PlainReporter{}
	stats := engine.CompletionStats{FilesCreated: 1}

	assert.NotPanics(t, func() {
		r.OnFileCreated("cmd/main.go")
		r.OnFileSkipped("redis.go", "condition false")
		r.OnHookStart("go mod tidy")
		r.OnHookOutput("downloading...")
		r.OnComplete(stats)
		r.OnError(assert.AnError)
	})
}
