package engine

type CompletionStats struct {
	FilesCreated  int
	FilesSkipped  int
	FilesModified int
	DurationMs    int64
}

// Reporter adalah interface yang digunakan engine untuk melaporkan
// progress dan status ke lapisan presentasi (TUI atau JSON output).
type Reporter interface {
	OnFileCreated(path string)
	OnFileSkipped(path string, reason string)
	OnHookStart(command string)
	OnHookOutput(line string)
	OnComplete(stats CompletionStats)
	OnError(err error)
}
