package tui

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	"github.com/username/symphony/internal/blueprint"
	"github.com/username/symphony/internal/engine"
)

// renderTemplate melakukan text/template rendering pada string dengan context values.
func renderTemplate(tmpl string, values map[string]any) string {
	t, err := template.New("msg").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return tmpl
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, values); err != nil {
		return tmpl
	}
	return buf.String()
}

// RenderCompletion merender hasil scaffolding menggunakan glamour.
func RenderCompletion(stats engine.CompletionStats, bp *blueprint.Blueprint) string {
	return RenderCompletionWithContext(stats, bp, nil)
}

// RenderCompletionWithContext merender completion screen dengan context values
// untuk keperluan template rendering di completion_message.
func RenderCompletionWithContext(stats engine.CompletionStats, bp *blueprint.Blueprint, values map[string]any) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s\n", StyleBrand.Render(IconDiamond), StyleHeader.Render("Scaffold Selesai!")))
	b.WriteString(Divider(54) + "\n\n")

	b.WriteString(fmt.Sprintf("  %s %d file dibuat  %s %d file dimodifikasi\n",
		StyleSuccess.Render(IconSuccess), stats.FilesCreated,
		StyleSuccess.Render(IconSuccess), stats.FilesModified,
	))

	b.WriteString(fmt.Sprintf("  %s Waktu eksekusi: %d ms\n", StyleSuccess.Render(IconSuccess), stats.DurationMs))

	if bp.CompletionMessage != "" {
		b.WriteString(Divider(54) + "\n\n")

		// Render template expressions dulu ({{.OUTPUT_DIR}} dll)
		msg := bp.CompletionMessage
		if values != nil {
			msg = renderTemplate(msg, values)
		}

		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err == nil {
			rendered, err := renderer.Render(msg)
			if err == nil {
				b.WriteString(rendered)
			} else {
				b.WriteString(msg + "\n")
			}
		} else {
			b.WriteString(msg + "\n")
		}
	}

	b.WriteString(Divider(54) + "\n")
	b.WriteString("  ✨ \033[1mHappy coding!\033[0m\n")

	return b.String()
}
