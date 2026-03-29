$ErrorActionPreference = "Stop"

function Make-Go-File {
    param([string]$Path)
    $dir = Split-Path $Path
    if (!(Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    $pkg = Split-Path $dir -Leaf
    $content = @"
// Package $pkg
// Tanggung jawab file ini akan diimplementasikan pada fase berikutnya.
package $pkg

// TODO: Implementasi di Phase selanjutnya
"@
    Set-Content -Path $Path -Value $content -Encoding UTF8
}

$files = @(
    "internal/engine/engine.go", "internal/engine/context.go", "internal/engine/walker.go", "internal/engine/renderer.go", "internal/engine/writer.go", "internal/engine/hooks.go",
    "internal/blueprint/parser.go", "internal/blueprint/validator.go", "internal/blueprint/schema.go", "internal/blueprint/resolver.go",
    "internal/ast/injector_interface.go", "internal/ast/go_injector.go", "internal/ast/ts_injector.go",
    "internal/remote/fetcher.go", "internal/remote/github.go", "internal/remote/cache.go",
    "internal/tui/prompt.go", "internal/tui/select.go", "internal/tui/multiselect.go", "internal/tui/input.go", "internal/tui/confirm.go", "internal/tui/progress.go", "internal/tui/summary.go",
    "internal/lock/writer.go", "internal/lock/reader.go",
    "internal/config/config.go",
    "pkg/expr/eval.go"
)

foreach ($f in $files) {
    Make-Go-File -Path $f
}

New-Item -ItemType Directory -Force -Path "testdata/templates/simple-go" | Out-Null
Set-Content -Path "testdata/templates/simple-go/template.yaml" -Value "# template.yaml kosong dengan komentar placeholder" -Encoding UTF8
New-Item -ItemType Directory -Force -Path "testdata/templates/hexagonal-go" | Out-Null
Set-Content -Path "testdata/templates/hexagonal-go/template.yaml" -Value "# template.yaml kosong dengan komentar placeholder" -Encoding UTF8

New-Item -ItemType Directory -Force -Path "docs" | Out-Null
Set-Content -Path "docs/blueprint-spec.md" -Value "<!-- TODO: Document blueprint spec -->" -Encoding UTF8
Set-Content -Path "docs/contributing.md" -Value "<!-- TODO: Document contribution guidelines -->" -Encoding UTF8
