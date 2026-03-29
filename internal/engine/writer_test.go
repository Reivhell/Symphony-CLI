package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile_PathTraversal_ReturnsError(t *testing.T) {
	out := t.TempDir()
	ctx := &EngineContext{
		OutputDir: out,
		DryRun:    false,
		YesAll:    true,
	}

	dangerous := []string{
		filepath.Join(out, "..", "..", "etc", "passwd"),
		filepath.Join(out, "normal", "..", "..", "outside"),
	}
	if os.PathSeparator == '\\' {
		dangerous = append(dangerous, filepath.Join(out, "..\\..\\Windows\\System32"))
	}

	for _, target := range dangerous {
		t.Run(target, func(t *testing.T) {
			err := WriteFile(target, "x", ctx)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrPathTraversal)
		})
	}
}

func TestSanitizePath_AllowsInsideOutput(t *testing.T) {
	out := t.TempDir()
	inside := filepath.Join(out, "sub", "file.txt")
	got, err := SanitizePath(out, inside)
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(got))
}

func TestWriteFile_DryRun_DoesNotWriteToDisk(t *testing.T) {
	out := t.TempDir()
	target := filepath.Join(out, "a", "b.txt")
	ctx := &EngineContext{
		OutputDir: out,
		DryRun:    true,
		YesAll:    true,
	}
	err := WriteFile(target, "hello", ctx)
	require.NoError(t, err)
	_, statErr := os.Stat(target)
	assert.Error(t, statErr)
}

func TestWriteFile_ParentDirCreated_WhenNotExists(t *testing.T) {
	out := t.TempDir()
	target := filepath.Join(out, "nested", "deep", "f.txt")
	ctx := &EngineContext{
		OutputDir: out,
		DryRun:    false,
		YesAll:    true,
	}
	require.NoError(t, WriteFile(target, "ok", ctx))
	b, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(b))
}
