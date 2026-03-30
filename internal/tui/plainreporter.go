package tui

import (
	"fmt"
	"os"

	"github.com/Reivhell/symphony/internal/engine"
)

type PlainReporter struct {
	TotalFiles int
	current    int
}

func NewPlainReporter(totalFiles int) *PlainReporter {
	return &PlainReporter{TotalFiles: totalFiles}
}

func (r *PlainReporter) OnFileCreated(path string) {
	r.current++
	fmt.Printf("[%d/%d] Created %s\n", r.current, r.TotalFiles, path)
}

func (r *PlainReporter) OnFileSkipped(path string, reason string) {
	fmt.Printf("Skipped %s (%s)\n", path, reason)
}

func (r *PlainReporter) OnHookStart(command string) {
	fmt.Printf("Running hook: %s\n", command)
}

func (r *PlainReporter) OnHookOutput(line string) {
	fmt.Printf("  %s\n", line)
}

func (r *PlainReporter) OnComplete(stats engine.CompletionStats) {
	fmt.Printf("\nScaffolding complete in %d ms.\nCreated: %d, Modified: %d, Skipped: %d.\n",
		stats.DurationMs, stats.FilesCreated, stats.FilesModified, stats.FilesSkipped)
}

func (r *PlainReporter) OnError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
