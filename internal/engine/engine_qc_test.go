package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Reivhell/symphony/internal/blueprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_RollbackOnInterrupt(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	tmplDir := filepath.Join(repoRoot, "testdata", "templates", "simple-go")

	bp, err := blueprint.Parse(tmplDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	engCtx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "demo",
			"USE_REDIS":    false,
		},
		SourceDir:    tmplDir,
		OutputDir:    outDir,
		DryRun:       false,
		NoHooks:      true,
		YesAll:       true,
		Format:       "human",
		WriteSession: NewWriteSession(),
		ExecCtx:      ctx,
		Meta: BlueprintMeta{
			Name: bp.Name, Version: bp.Version, Source: tmplDir, Commit: "local",
		},
	}

	eng := New(bp, engCtx)
	actions, err := eng.Prepare()
	require.NoError(t, err)

	err = eng.Execute(ctx, actions)
	require.ErrorIs(t, err, context.Canceled)
}

func TestEngine_ConditionalActions_SkipsCorrectly(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	tmplDir := filepath.Join(repoRoot, "testdata", "templates", "simple-go")

	bp, err := blueprint.Parse(tmplDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	engCtx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "demo",
			"USE_REDIS":    false,
		},
		SourceDir: tmplDir,
		OutputDir: outDir,
		DryRun:    false,
		NoHooks:   true,
		YesAll:    true,
		Format:    "human",
		Meta:      BlueprintMeta{Name: bp.Name, Version: bp.Version, Source: tmplDir, Commit: "local"},
	}
	eng := New(bp, engCtx)
	require.NoError(t, eng.Run())

	_, statErr := os.Stat(filepath.Join(outDir, "README.md"))
	assert.Error(t, statErr, "README render is conditional on USE_REDIS")
}

func TestEngine_HookFailure_ReturnsError(t *testing.T) {
	tmp := t.TempDir()
	hookCmd := `sh -c "exit 1"`
	if runtime.GOOS == "windows" {
		hookCmd = "exit 1"
	}
	yaml := fmt.Sprintf(`schema_version: "2"
name: "hookfail"
version: "1.0"
prompts:
  - id: X
    question: "x"
    type: input
    default: "y"
actions:
  - type: render
    source: "f.tmpl"
    target: "out.txt"
  - type: shell
    command: %q
`, hookCmd)
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(yaml), 0644))
	tmpl := "hello"
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "f.tmpl"), []byte(tmpl), 0644))

	bp, err := blueprint.Parse(tmp)
	require.NoError(t, err)

	out := t.TempDir()
	engCtx := &EngineContext{
		Values:    map[string]any{"X": "y"},
		SourceDir: tmp,
		OutputDir: out,
		DryRun:    false,
		NoHooks:   false,
		YesAll:    true,
		Format:    "human",
		Meta:      BlueprintMeta{Name: "hookfail", Version: "1.0", Source: tmp, Commit: "local"},
	}
	eng := New(bp, engCtx)
	err = eng.Run()
	require.Error(t, err)
}

func TestEngine_FullScaffold_WithAllFeatures(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	tmplDir := filepath.Join(repoRoot, "testdata", "templates", "hexagonal-go")

	bp, err := blueprint.Parse(tmplDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	engCtx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "e2eproj",
			"USE_CACHE":    true,
		},
		SourceDir: tmplDir,
		OutputDir: outDir,
		DryRun:    false,
		NoHooks:   true,
		YesAll:    true,
		Format:    "human",
		Meta:      BlueprintMeta{Name: bp.Name, Version: bp.Version, Source: tmplDir, Commit: "local"},
	}
	eng := New(bp, engCtx)
	require.NoError(t, eng.Run())

	b, err := os.ReadFile(filepath.Join(outDir, "cmd", "e2eproj", "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "e2eproj")
}
