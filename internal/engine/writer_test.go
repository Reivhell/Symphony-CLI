package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestWriter_BoundedConcurrency(t *testing.T) {
	var peak int32
	var cur int32
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = acquireWriteSlot()
			n := atomic.AddInt32(&cur, 1)
			for {
				o := atomic.LoadInt32(&peak)
				if n <= o {
					break
				}
				if atomic.CompareAndSwapInt32(&peak, o, n) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt32(&cur, -1)
			releaseWriteSlot()
		}(i)
	}
	wg.Wait()
	assert.LessOrEqual(t, peak, int32(10), "semaphore allows at most 10 concurrent critical sections")
}

func TestWriteFile_ConcurrentUniquePaths(t *testing.T) {
	out := t.TempDir()
	ctx := &EngineContext{
		OutputDir: out,
		DryRun:    false,
		YesAll:    true,
	}
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := filepath.Join(out, fmt.Sprintf("d%d", i), "x.txt")
			require.NoError(t, WriteFile(p, "ok", ctx))
		}(i)
	}
	wg.Wait()
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
