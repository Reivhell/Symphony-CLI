package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPluginForSource(t *testing.T) {
	plugins := []PluginRenderer{
		{Name: "a", Handles: []string{"*.prisma"}},
		{Name: "b", Handles: []string{"*.proto"}},
	}
	assert.Nil(t, findPluginForSource("/tmp/x.txt", plugins))
	p := findPluginForSource("/tmp/schema.prisma", plugins)
	require.NotNil(t, p)
	assert.Equal(t, "a", p.Name)
}

func TestResolvePluginExecutable_rejectsDotDot(t *testing.T) {
	_, err := resolvePluginExecutable("../bin/x", "/src")
	assert.Error(t, err)
}

func TestPluginRenderer_Render_echoPlugin(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	pluginSrcDir := filepath.Join(wd, "testdata", "plugin_echo")
	out := filepath.Join(t.TempDir(), "plugin_echo")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = pluginSrcDir
	outBytes, err := cmd.CombinedOutput()
	require.NoError(t, err, string(outBytes))

	if runtime.GOOS != "windows" {
		require.NoError(t, os.Chmod(out, 0755))
	}

	tmp := t.TempDir()
	ctx := &EngineContext{
		Values:    map[string]any{"K": "v"},
		SourceDir: tmp,
		OutputDir: tmp,
	}
	rel := filepath.Base(out)
	// Place executable next to template root
	require.NoError(t, os.Rename(out, filepath.Join(tmp, rel)))

	p := &PluginRenderer{
		Name:       "echo",
		Executable: rel,
		Handles:    []string{"*.custom"},
	}

	src := filepath.Join(tmp, "file.custom")
	require.NoError(t, os.WriteFile(src, []byte("BODY"), 0644))

	got, err := p.Render(ctx, src, filepath.Join(tmp, "out.custom"), "BODY")
	require.NoError(t, err)
	assert.Equal(t, "PLUGIN:BODY", got)
}
