package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/username/symphony/internal/blueprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineRun_DryRun(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	tmplDir := filepath.Join(repoRoot, "testdata", "templates", "simple-go")

	bp, err := blueprint.Parse(tmplDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	ctx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "demo",
			"USE_REDIS":    false,
		},
		SourceDir: tmplDir,
		OutputDir: outDir,
		DryRun:    true,
		YesAll:    true,
		Format:    "human",
		Meta: BlueprintMeta{
			Name:    bp.Name,
			Version: bp.Version,
			Source:  tmplDir,
			Commit:  "local",
		},
	}

	eng := New(bp, ctx)
	err = eng.Run()
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(outDir, "symphony.lock"))
	assert.Error(t, err, "dry-run tidak boleh menulis symphony.lock")
}

func TestEngineRun_WritesFilesAndLock(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	tmplDir := filepath.Join(repoRoot, "testdata", "templates", "simple-go")

	bp, err := blueprint.Parse(tmplDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	ctx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "demo",
			"USE_REDIS":    false,
		},
		SourceDir: tmplDir,
		OutputDir: outDir,
		DryRun:    false,
		YesAll:    true,
		NoHooks:   true,
		Format:    "human",
		Meta: BlueprintMeta{
			Name:    bp.Name,
			Version: bp.Version,
			Source:  tmplDir,
			Commit:  "local",
		},
	}

	eng := New(bp, ctx)
	err = eng.Run()
	require.NoError(t, err)

	mainPath := filepath.Join(outDir, "cmd", "demo", "main.go")
	_, err = os.Stat(mainPath)
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(outDir, "symphony.lock"))
	assert.NoError(t, err)
}
