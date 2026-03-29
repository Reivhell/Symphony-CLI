package engine

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var funcMap = template.FuncMap{
	"toLower":   strings.ToLower,
	"toUpper":   strings.ToUpper,
	"toCamel":   toCamel,
	"toSnake":   toSnake,
	"hasPrefix": strings.HasPrefix,
	"now":       func() string { return time.Now().Format(time.RFC3339) },
}

func toCamel(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.English, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}

func toSnake(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return strings.ToLower(s)
}

func RenderFile(sourcePath, targetPath string, ctx *EngineContext) (string, error) {
	contentBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}
	contentStr := string(contentBytes)

	if p := findPluginForSource(sourcePath, ctx.Plugins); p != nil {
		return p.Render(ctx, sourcePath, targetPath, contentStr)
	}

	if !strings.HasSuffix(sourcePath, ".tmpl") {
		return contentStr, nil
	}

	return RenderString(contentStr, ctx)
}

func RenderString(tmplStr string, ctx *EngineContext) (string, error) {
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	data := templateDataForRender(ctx)

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// templateDataForRender builds template data. User-controlled string values are passed
// through escapeTemplateMeta so delimiter sequences from prompts cannot influence parsing
// in edge cases (defense in depth; text/template executes in one pass).
func templateDataForRender(ctx *EngineContext) map[string]any {
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

func escapeTemplateMeta(s string) string {
	s = strings.ReplaceAll(s, "{{", "\u200b{{\u200b")
	s = strings.ReplaceAll(s, "}}", "\u200b}}\u200b")
	return s
}
