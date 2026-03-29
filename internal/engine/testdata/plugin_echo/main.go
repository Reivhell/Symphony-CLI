// Test helper: reads PluginRequest JSON from stdin, writes PluginResponse JSON to stdout.
package main

import (
	"encoding/json"
	"os"
)

type pluginRequest struct {
	Context     map[string]any `json:"context"`
	FileContent string         `json:"file_content"`
	SourcePath  string         `json:"source_path"`
	TargetPath  string         `json:"target_path"`
}

type pluginResponse struct {
	RenderedContent string `json:"rendered_content"`
	Error           string `json:"error,omitempty"`
}

func main() {
	var req pluginRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(pluginResponse{Error: err.Error()})
		return
	}
	out := "PLUGIN:" + req.FileContent
	_ = json.NewEncoder(os.Stdout).Encode(pluginResponse{RenderedContent: out})
}
