package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Reivhell/symphony/internal/blueprint"
)

const pluginExecTimeout = 30 * time.Second

// PluginRenderer runs an external binary to render a file (JSON over stdin/stdout).
type PluginRenderer struct {
	Name       string
	Executable string
	Handles    []string
}

// PluginRequest is sent to the plugin process on stdin.
type PluginRequest struct {
	Context     map[string]any `json:"context"`
	FileContent string         `json:"file_content"`
	SourcePath  string         `json:"source_path"`
	TargetPath  string         `json:"target_path"`
}

// PluginResponse is read from the plugin process stdout.
type PluginResponse struct {
	RenderedContent string `json:"rendered_content"`
	Error           string `json:"error,omitempty"`
}

// PluginRenderersFromBlueprint builds engine plugin descriptors from a blueprint.
func PluginRenderersFromBlueprint(bp *blueprint.Blueprint) []PluginRenderer {
	if bp == nil || len(bp.Plugins) == 0 {
		return nil
	}
	out := make([]PluginRenderer, 0, len(bp.Plugins))
	for _, p := range bp.Plugins {
		out = append(out, PluginRenderer{
			Name:       p.Name,
			Executable: p.Executable,
			Handles:    p.Handles,
		})
	}
	return out
}

// findPluginForSource returns the first registered plugin whose handle glob matches the source file base name.
func findPluginForSource(sourcePath string, plugins []PluginRenderer) *PluginRenderer {
	base := filepath.Base(sourcePath)
	for i := range plugins {
		p := &plugins[i]
		for _, pattern := range p.Handles {
			ok, err := filepath.Match(pattern, base)
			if err == nil && ok {
				return p
			}
		}
	}
	return nil
}

func resolvePluginExecutable(executable, sourceDir string) (string, error) {
	if strings.TrimSpace(executable) == "" {
		return "", fmt.Errorf("plugin executable is empty")
	}
	if strings.Contains(executable, "..") {
		return "", fmt.Errorf("plugin executable must not contain '..': %q", executable)
	}

	var resolved string
	if filepath.IsAbs(executable) {
		resolved = filepath.Clean(executable)
	} else {
		resolved = filepath.Clean(filepath.Join(sourceDir, executable))
	}

	if strings.Contains(resolved, "..") {
		return "", fmt.Errorf("invalid plugin executable path after resolution: %q", resolved)
	}

	fi, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("plugin executable not accessible: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("plugin path is not a regular file: %s", resolved)
	}

	if runtime.GOOS != "windows" {
		if fi.Mode()&0111 == 0 {
			return "", fmt.Errorf("plugin file is not executable: %s", resolved)
		}
	}

	return resolved, nil
}

func pluginContextPayload(ctx *EngineContext) map[string]any {
	data := make(map[string]any)
	if ctx.Values != nil {
		for k, v := range ctx.Values {
			if s, ok := v.(string); ok {
				data[k] = escapeTemplateMeta(s)
			} else {
				data[k] = v
			}
		}
	}
	data["OUTPUT_DIR"] = ctx.OutputDir
	data["SOURCE_DIR"] = ctx.SourceDir
	return data
}

// Render runs the plugin binary and returns rendered output.
func (p *PluginRenderer) Render(ctx *EngineContext, sourcePath, targetPath, sourceContent string) (string, error) {
	if p == nil {
		return "", fmt.Errorf("nil plugin renderer")
	}
	exe, err := resolvePluginExecutable(p.Executable, ctx.SourceDir)
	if err != nil {
		return "", fmt.Errorf("plugin %q: %w", p.Name, err)
	}

	req := PluginRequest{
		Context:     pluginContextPayload(ctx),
		FileContent: sourceContent,
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("plugin %q: marshal request: %w", p.Name, err)
	}

	execCtx, cancel := context.WithTimeout(context.Background(), pluginExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, exe)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("plugin %q: timed out after %v", p.Name, pluginExecTimeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("plugin %q: %w: %s", p.Name, err, msg)
		}
		return "", fmt.Errorf("plugin %q: %w", p.Name, err)
	}

	var resp PluginResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return "", fmt.Errorf("plugin %q: invalid JSON on stdout: %w", p.Name, err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("plugin %q: %s", p.Name, resp.Error)
	}
	return resp.RenderedContent, nil
}
