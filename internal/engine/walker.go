package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/username/symphony/internal/blueprint"
	"github.com/username/symphony/pkg/expr"
)

type ResolvedAction struct {
	Original   blueprint.Action
	SourcePath string
	TargetPath string
	ShouldRun  bool
	SkipReason string
}

func Walk(bp *blueprint.Blueprint, ctx *EngineContext) ([]ResolvedAction, error) {
	var resolved []ResolvedAction

	for _, action := range bp.Actions {
		res := ResolvedAction{
			Original: action,
		}

		switch action.Type {
		case "render":
			if strings.TrimSpace(action.Source) == "" || strings.TrimSpace(action.Target) == "" {
				return nil, fmt.Errorf("action render harus memiliki source dan target yang tidak kosong")
			}
		case "shell":
			if strings.TrimSpace(action.Command) == "" {
				return nil, fmt.Errorf("action shell harus memiliki command")
			}
		case "ast-inject":
			if strings.TrimSpace(action.Target) == "" {
				return nil, fmt.Errorf("action ast-inject harus memiliki target")
			}
		}

		if action.If != "" {
			shouldRun, err := expr.Evaluate(action.If, ctx.Values)
			if err != nil {
				return nil, fmt.Errorf("evaluasi kondisi '%s' gagal: %w", action.If, err)
			}
			res.ShouldRun = shouldRun
			if !shouldRun {
				res.SkipReason = "Kondisi if menghasilkan false: " + action.If
			}
		} else {
			res.ShouldRun = true
		}

		if action.Type == "render" {
			res.SourcePath = filepath.Join(ctx.SourceDir, action.Source)
			
			targetStr, err := RenderString(action.Target, ctx)
			if err != nil {
				return nil, fmt.Errorf("gagal me-render target path '%s': %w", action.Target, err)
			}
			res.TargetPath = filepath.Join(ctx.OutputDir, targetStr)
		}

		resolved = append(resolved, res)
	}

	return resolved, nil
}
