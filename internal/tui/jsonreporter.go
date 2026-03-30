package tui

import (
	"encoding/json"
	"fmt"

	"github.com/Reivhell/symphony/internal/engine"
)

type JSONReporter struct {
	TotalFiles int
	current    int
}

func NewJSONReporter(totalFiles int) *JSONReporter {
	return &JSONReporter{TotalFiles: totalFiles}
}

func (r *JSONReporter) OnFileCreated(path string) {
	r.current++
	r.emit(map[string]any{
		"type":    "progress",
		"current": r.current,
		"total":   r.TotalFiles,
		"file":    path,
		"action":  "CREATE",
	})
}

func (r *JSONReporter) OnFileSkipped(path string, reason string) {
	r.emit(map[string]any{
		"type":   "file_skip",
		"file":   path,
		"reason": reason,
	})
}

func (r *JSONReporter) OnHookStart(command string) {
	r.emit(map[string]any{
		"type":    "hook_start",
		"command": command,
	})
}

func (r *JSONReporter) OnHookOutput(line string) {
	r.emit(map[string]any{
		"type": "hook_output",
		"line": line,
	})
}

func (r *JSONReporter) OnComplete(stats engine.CompletionStats) {
	r.emit(map[string]any{
		"type":          "complete",
		"files_created": stats.FilesCreated,
		"files_skipped": stats.FilesSkipped,
		"duration_ms":   stats.DurationMs,
	})
}

func (r *JSONReporter) OnError(err error) {
	r.emit(map[string]any{
		"type":    "error",
		"message": err.Error(),
	})
}

func (r *JSONReporter) emit(data map[string]any) {
	b, _ := json.Marshal(data)
	fmt.Println(string(b))
}

// RenderJSONSummary untuk print `--format json` di bagian summary
func RenderJSONSummary(actions []engine.ResolvedAction) {
	var formatted []map[string]any

	for _, a := range actions {
		item := map[string]any{
			"path": a.TargetPath,
		}
		if !a.ShouldRun {
			item["action"] = "SKIP"
			item["reason"] = a.SkipReason
			src := a.Original.Source
			if src == "" {
				src = a.Original.Target
			}
			item["path"] = src
		} else {
			switch a.Original.Type {
			case "shell":
				item["action"] = "HOOK"
				item["command"] = a.Original.Command
			case "ast-inject":
				item["action"] = "AST-INJECT"
			default:
				item["action"] = "CREATE"
			}
		}
		formatted = append(formatted, item)
	}

	b, _ := json.Marshal(map[string]any{
		"type":    "dry_run_summary",
		"actions": formatted,
	})
	fmt.Println(string(b))
}
