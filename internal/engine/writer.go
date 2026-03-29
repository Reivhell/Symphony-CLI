package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/semaphore"
)

// ErrPathTraversal is returned when a target path would escape the output directory.
var ErrPathTraversal = errors.New("path escapes output directory")

// concurrentWriteLimit bounds parallel disk writes (max 10 concurrent).
var concurrentWriteLimit = semaphore.NewWeighted(10)

func acquireWriteSlot() error {
	return concurrentWriteLimit.Acquire(context.Background(), 1)
}

func releaseWriteSlot() {
	concurrentWriteLimit.Release(1)
}

type WriteResult struct {
	Path   string
	Action string // "created" | "skipped" | "modified" | "dry-run"
	Error  error
}

// SanitizePath returns an absolute path for targetPath and verifies it stays within outputDir.
func SanitizePath(outputDir, targetPath string) (string, error) {
	outAbs, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}
	targAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}

	outAbs = filepath.Clean(outAbs)
	targAbs = filepath.Clean(targAbs)

	rel, err := filepath.Rel(outAbs, targAbs)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrPathTraversal, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	return targAbs, nil
}

// WriteFile writes content to targetPath under ctx.OutputDir.
// It rejects paths that escape the output directory (path traversal).
func WriteFile(targetPath string, content string, ctx *EngineContext) error {
	if ctx.DryRun {
		_, err := SanitizePath(ctx.OutputDir, targetPath)
		return err
	}

	safePath, err := SanitizePath(ctx.OutputDir, targetPath)
	if err != nil {
		return err
	}
	targetPath = safePath

	if err := writeFileWithSemaphore(targetPath, content, ctx); err != nil {
		return err
	}

	return nil
}

func writeFileWithSemaphore(targetPath string, content string, ctx *EngineContext) error {
	if err := acquireWriteSlot(); err != nil {
		return err
	}
	defer releaseWriteSlot()

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		if !ctx.YesAll {
			fmt.Printf("File exists: %s\nOverwrite? (y/N): ", targetPath)
			line, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("read confirmation: %w", err)
			}
			s := strings.TrimSpace(strings.ToLower(line))
			if s != "y" && s != "yes" {
				return fmt.Errorf("write cancelled: %s", targetPath)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat file %s: %w", targetPath, err)
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file %s: %w", targetPath, err)
	}

	if ctx.WriteSession != nil {
		ctx.WriteSession.Record(targetPath)
	}

	return nil
}
