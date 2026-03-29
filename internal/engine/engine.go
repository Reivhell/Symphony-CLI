package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/username/symphony/internal/ast"
	"github.com/username/symphony/internal/blueprint"
	"github.com/username/symphony/internal/lock"
	"github.com/username/symphony/internal/version"
)

type Engine struct {
	blueprint *blueprint.Blueprint
	ctx       *EngineContext
}

func New(bp *blueprint.Blueprint, ctx *EngineContext) *Engine {
	return &Engine{
		blueprint: bp,
		ctx:       ctx,
	}
}

// Prepare melakukan validasi dan resolve semua action sebelum dieksekusi
func (e *Engine) Prepare() ([]ResolvedAction, error) {
	if err := ValidateInputs(e.blueprint, e.ctx); err != nil {
		return nil, err
	}

	actions, err := Walk(e.blueprint, e.ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengevaluasi template: %w", err)
	}

	return actions, nil
}

// Execute menjalankan proses pembuatan file dan eksekusi hook berdasarkan list aksi yang sudah disusun.
func (e *Engine) Execute(execCtx context.Context, actions []ResolvedAction) error {
	if execCtx == nil {
		execCtx = context.Background()
	}

	var stats CompletionStats
	startTime := time.Now()

	for _, a := range actions {
		select {
		case <-execCtx.Done():
			if e.ctx.WriteSession != nil {
				_ = e.ctx.WriteSession.Rollback()
			}
			return execCtx.Err()
		default:
		}

		if !a.ShouldRun {
			if e.ctx.Reporter != nil {
				e.ctx.Reporter.OnFileSkipped(a.TargetPath, a.SkipReason)
			}
			stats.FilesSkipped++
			continue
		}
		switch a.Original.Type {
		case "render":
			if e.ctx.DryRun {
				continue
			}
			content, err := RenderFile(a.SourcePath, a.TargetPath, e.ctx)
			if err != nil {
				if e.ctx.Reporter != nil {
					e.ctx.Reporter.OnError(fmt.Errorf("gagal me-render %s: %w", a.SourcePath, err))
				}
				return fmt.Errorf("gagal me-render %s: %w", a.SourcePath, err)
			}
			if err := WriteFile(a.TargetPath, content, e.ctx); err != nil {
				if e.ctx.Reporter != nil {
					e.ctx.Reporter.OnError(fmt.Errorf("gagal menulis ke %s: %w", a.TargetPath, err))
				}
				return fmt.Errorf("gagal menulis ke %s: %w", a.TargetPath, err)
			}
			if e.ctx.Reporter != nil {
				e.ctx.Reporter.OnFileCreated(a.TargetPath)
			}
			stats.FilesCreated++
		case "shell":
			continue
		case "ast-inject":
			if e.ctx.DryRun {
				continue
			}
			
			var inj ast.Injector
			goInj := &ast.GoInjector{}
			if goInj.CanHandle(a.TargetPath) {
				inj = goInj
			} else {
				inj = &ast.AnchorInjector{}
			}

			renderedContent, err := RenderString(a.Original.Content, e.ctx)
			if err != nil {
				return fmt.Errorf("gagal me-render ast-inject template: %w", err)
			}

			actCopy := a.Original
			actCopy.Content = renderedContent

			if err := inj.Inject(a.TargetPath, actCopy); err != nil {
				if e.ctx.Reporter != nil {
					e.ctx.Reporter.OnError(fmt.Errorf("gagal di-inject ke %s: %w", a.TargetPath, err))
				}
				return fmt.Errorf("gagal injeksi %s: %w", a.TargetPath, err)
			}

			if e.ctx.Reporter != nil {
				e.ctx.Reporter.OnFileCreated(a.TargetPath) // Di UI reporter terhitung "dibuat" (modified)
			}
			stats.FilesCreated++
		default:
			return fmt.Errorf("action type tidak dikenal: %s", a.Original.Type)
		}
	}

	if !e.ctx.NoHooks && !e.ctx.DryRun {
		for _, a := range actions {
			select {
			case <-execCtx.Done():
				if e.ctx.WriteSession != nil {
					_ = e.ctx.WriteSession.Rollback()
				}
				return execCtx.Err()
			default:
			}
			if a.ShouldRun && a.Original.Type == "shell" {
				if err := RunHook(execCtx, a.Original, e.ctx); err != nil {
					if e.ctx.Reporter != nil {
						e.ctx.Reporter.OnError(err)
					}
					if errors.Is(err, context.Canceled) {
						if e.ctx.WriteSession != nil {
							_ = e.ctx.WriteSession.Rollback()
						}
					}
					return err
				}
			}
		}
	}

	if !e.ctx.DryRun {
		lf := &lock.LockFile{
			SymphonyVersion: version.Version,
			GeneratedAt:     time.Now(),
			Template: lock.TemplateLockInfo{
				Source:  e.ctx.Meta.Source,
				Version: e.ctx.Meta.Version,
				Commit:  e.ctx.Meta.Commit,
			},
			Inputs:         e.ctx.Values,
			OutputChecksum: "",
		}
		if err := lock.Write(lf, e.ctx.OutputDir); err != nil {
			if e.ctx.Reporter != nil {
				e.ctx.Reporter.OnError(fmt.Errorf("gagal membuat lock file: %w", err))
			}
		}
	}

	stats.DurationMs = time.Since(startTime).Milliseconds()
	if e.ctx.Reporter != nil {
		e.ctx.Reporter.OnComplete(stats)
	}

	return nil
}

// Run adalah convenience method yang menggabungkan Prepare() dan Execute().
// Berguna untuk use case sederhana di test dan CLI.
func (e *Engine) Run() error {
	actions, err := e.Prepare()
	if err != nil {
		return err
	}
	return e.Execute(context.Background(), actions)
}
