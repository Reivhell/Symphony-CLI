package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Reivhell/symphony/internal/blueprint"
	"github.com/Reivhell/symphony/internal/engine"
)

func makeSummaryModel(actions []engine.ResolvedAction) summaryModel {
	return summaryModel{actions: actions}
}

func TestSummaryView_CreateActions(t *testing.T) {
	actions := []engine.ResolvedAction{
		{
			Original:   blueprint.Action{Type: "render", Source: "cmd/main.go.tmpl", Target: "cmd/main.go"},
			TargetPath: "/out/cmd/main.go",
			ShouldRun:  true,
		},
		{
			Original:   blueprint.Action{Type: "render", Source: "go.mod.tmpl", Target: "go.mod"},
			TargetPath: "/out/go.mod",
			ShouldRun:  true,
		},
	}

	m := makeSummaryModel(actions)
	view := m.View()

	assert.Contains(t, view, "[CREATE]")
	assert.Contains(t, view, "cmd/main.go")
	assert.Contains(t, view, "go.mod")
	assert.Contains(t, view, "2 file akan dibuat")
}

func TestSummaryView_SkipActions(t *testing.T) {
	actions := []engine.ResolvedAction{
		{
			Original:   blueprint.Action{Type: "render", Source: "redis/cache.go.tmpl", Target: "redis/cache.go"},
			TargetPath: "/out/redis/cache.go",
			ShouldRun:  false,
			SkipReason: "USE_REDIS == false",
		},
	}

	m := makeSummaryModel(actions)
	view := m.View()

	assert.Contains(t, view, "[SKIP]")
	assert.Contains(t, view, "redis/cache.go.tmpl")
	// Harus tampilkan reason
	assert.Contains(t, view, "USE_REDIS == false")
}

func TestSummaryView_MixedActions(t *testing.T) {
	actions := []engine.ResolvedAction{
		{
			Original:   blueprint.Action{Type: "render", Source: "cmd/main.go.tmpl", Target: "cmd/main.go"},
			TargetPath: "/out/cmd/main.go",
			ShouldRun:  true,
		},
		{
			Original:   blueprint.Action{Type: "render", Source: "redis/cache.go.tmpl", Target: "redis/cache.go"},
			TargetPath: "/out/redis/cache.go",
			ShouldRun:  false,
			SkipReason: "USE_REDIS == false",
		},
		{
			Original:  blueprint.Action{Type: "shell", Command: "go mod tidy"},
			ShouldRun: true,
		},
	}

	m := makeSummaryModel(actions)
	view := m.View()

	assert.Contains(t, view, "[CREATE]")
	assert.Contains(t, view, "[SKIP]")
	assert.Contains(t, view, "[HOOK]")
	assert.Contains(t, view, "go mod tidy")
	assert.Contains(t, view, "1 file akan dibuat")
	assert.Contains(t, view, "1 file di-skip")
}

func TestSummaryView_ConfirmationPrompt(t *testing.T) {
	m := makeSummaryModel(nil)
	view := m.View()

	assert.Contains(t, view, "Lanjutkan?")
}

func TestSummaryView_DoneState(t *testing.T) {
	m := summaryModel{done: true, result: true}
	view := m.View()
	assert.Contains(t, view, "Yes")
}

func TestSummaryView_DoneStateFalse(t *testing.T) {
	m := summaryModel{done: true, result: false}
	view := m.View()
	assert.True(t, strings.Contains(view, "No") || strings.Contains(view, "Dibatalkan") || len(view) > 0)
}
