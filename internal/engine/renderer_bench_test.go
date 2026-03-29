package engine

import (
	"strings"
	"testing"
)

func BenchmarkRenderString_Simple(b *testing.B) {
	ctx := &EngineContext{
		Values: map[string]any{
			"A": "one", "B": "two", "C": "three", "D": "four", "E": "five",
		},
		OutputDir: "/out",
		SourceDir: "/src",
	}
	tmpl := `{{.A}}{{.B}}{{.C}}{{.D}}{{.E}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderString(tmpl, ctx)
	}
}

func BenchmarkRenderString_Complex(b *testing.B) {
	ctx := &EngineContext{
		Values: map[string]any{
			"V00": "a", "V01": "b", "V02": "c", "V03": "d", "V04": "e",
			"V05": "f", "V06": "g", "V07": "h", "V08": "i", "V09": "j",
			"V10": "k", "V11": "l", "V12": "m", "V13": "n", "V14": "o",
			"V15": "p", "V16": "q", "V17": "r", "V18": "s", "V19": "t",
		},
		OutputDir: "/out",
		SourceDir: "/src",
	}
	tmpl := `{{if eq .V00 "a"}}{{.V01}}{{end}}{{if .V02}}{{.V03}}{{end}}{{.V04}}{{.V05}}{{.V06}}{{.V07}}{{.V08}}{{.V09}}` +
		strings.Repeat(`{{.V10}}{{.V11}}{{.V12}}{{.V13}}{{.V14}}`, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderString(tmpl, ctx)
	}
}

func BenchmarkRenderString_Large(b *testing.B) {
	body := strings.Repeat("line {{.K}}\n", 500)
	ctx := &EngineContext{
		Values:    map[string]any{"K": "v"},
		OutputDir: "/out",
		SourceDir: "/src",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderString(body, ctx)
	}
}
