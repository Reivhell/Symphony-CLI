package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/symphony/internal/blueprint"
	"github.com/username/symphony/internal/engine"
)

func TestRenderCompletion_BasicOutput(t *testing.T) {
	stats := engine.CompletionStats{
		FilesCreated:  5,
		FilesModified: 1,
		FilesSkipped:  2,
		DurationMs:    342,
	}
	bp := &blueprint.Blueprint{
		Name:              "Test Template",
		CompletionMessage: "",
	}

	out := RenderCompletion(stats, bp)

	assert.Contains(t, out, "Scaffold Selesai!")
	assert.Contains(t, out, "5 file dibuat")
	assert.Contains(t, out, "342 ms")
	assert.Contains(t, out, "Happy coding")
}

func TestRenderCompletion_WithMarkdownMessage(t *testing.T) {
	stats := engine.CompletionStats{
		FilesCreated: 3,
		DurationMs:   100,
	}
	bp := &blueprint.Blueprint{
		Name:              "Test Template",
		CompletionMessage: "## Langkah Selanjutnya\n\nJalankan `go run ./cmd/main.go`",
	}

	out := RenderCompletion(stats, bp)

	assert.Contains(t, out, "Scaffold Selesai!")
	// Markdown harus dirender — isi teks tetap ada
	assert.True(t, strings.Contains(out, "Langkah Selanjutnya") || strings.Contains(out, "langkah"))
	assert.Contains(t, out, "Happy coding")
}

func TestRenderCompletion_NoCompletionMessage(t *testing.T) {
	stats := engine.CompletionStats{FilesCreated: 1}
	bp := &blueprint.Blueprint{CompletionMessage: ""}

	out := RenderCompletion(stats, bp)

	assert.Contains(t, out, "Scaffold Selesai!")
	assert.NotContains(t, out, "undefined") // tidak ada template error
}

func TestRenderCompletion_ZeroStats(t *testing.T) {
	stats := engine.CompletionStats{}
	bp := &blueprint.Blueprint{}

	assert.NotPanics(t, func() {
		out := RenderCompletion(stats, bp)
		assert.NotEmpty(t, out)
	})
}
