package engine

import "context"

type EngineContext struct {
	// Values adalah map dari semua jawaban prompt
	Values map[string]any

	// Meta berisi informasi tentang template
	Meta BlueprintMeta

	// SourceDir adalah path absolut ke direktori template
	SourceDir string

	// OutputDir adalah path absolut ke direktori output
	OutputDir string

	// DryRun jika true, tidak ada file yang ditulis ke disk
	DryRun bool

	// NoHooks jika true, post-scaffold hooks tidak dijalankan
	NoHooks bool

	// YesAll jika true, lewati konfirmasi overwrite dan prompt "Lanjutkan?"
	YesAll bool

	// Format adalah output format: "human" atau "json"
	Format string

	// Reporter adalah interface untuk melaporkan progress
	Reporter Reporter

	// Plugins are external renderers (first glob match wins). Optional.
	Plugins []PluginRenderer

	// ExecCtx is the CLI request context (cancellation). If nil, engine uses context.Background().
	ExecCtx context.Context

	// WriteSession tracks written files for rollback when ExecCtx is cancelled. Optional.
	WriteSession *WriteSession
}

type BlueprintMeta struct {
	Name    string
	Version string
	Source  string
	Commit  string
}

// Get mengambil nilai dari context
func (c *EngineContext) Get(key string) any {
	if c.Values == nil {
		return nil
	}
	return c.Values[key]
}

// GetString mengambil nilai sebagai string
func (c *EngineContext) GetString(key string) string {
	val := c.Get(key)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetBool mengambil nilai sebagai bool
func (c *EngineContext) GetBool(key string) bool {
	val := c.Get(key)
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
