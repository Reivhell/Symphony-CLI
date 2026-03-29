package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderString(t *testing.T) {
	ctx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "my-app",
			"USE_REDIS":    true,
		},
		OutputDir: "/out",
	}

	tmpl := `Project: {{.PROJECT_NAME}} - DIR: {{.OUTPUT_DIR}}`
	res, err := RenderString(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Project: my-app - DIR: /out", res)

	tmplIf := `{{if .USE_REDIS}}Redis On{{else}}Redis Off{{end}}`
	resIf, err := RenderString(tmplIf, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Redis On", resIf)

	tmplFunc := `{{toUpper .PROJECT_NAME}}`
	resFunc, err := RenderString(tmplFunc, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "MY-APP", resFunc)
}

func TestRenderString_TemplateInjection_UserValueDoesNotExpandSecrets(t *testing.T) {
	ctx := &EngineContext{
		Values: map[string]any{
			"PROJECT_NAME": "{{.SECRET_VAR}}",
			"SECRET_VAR":   "sensitive-data",
		},
		OutputDir: "/out",
		SourceDir: "/src",
	}
	// Only PROJECT_NAME is expanded; it must not reveal SECRET_VAR via nested evaluation.
	out, err := RenderString(`{{.PROJECT_NAME}}`, ctx)
	require.NoError(t, err)
	assert.NotContains(t, out, "sensitive-data")
}

func TestRenderFile(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := &EngineContext{
		Values: map[string]any{"NAME": "test"},
	}

	tmplPath := filepath.Join(tmpDir, "test.tmpl")
	err := os.WriteFile(tmplPath, []byte(`Hello {{.NAME}}`), 0644)
	assert.NoError(t, err)

	res, err := RenderFile(tmplPath, filepath.Join(tmpDir, "out.tmpl"), ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Hello test", res)

	notTmplPath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(notTmplPath, []byte(`Hello {{.NAME}}`), 0644)
	assert.NoError(t, err)

	resNotTmpl, err := RenderFile(notTmplPath, filepath.Join(tmpDir, "out.txt"), ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Hello {{.NAME}}", resNotTmpl)
}
