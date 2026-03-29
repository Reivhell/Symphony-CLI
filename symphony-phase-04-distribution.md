# Symphony CLI — Phase 04: Distribution & Ecosystem
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 4 dari 5  
> **Tujuan:** Siapkan Symphony untuk distribusi publik — plugin system, multi-platform binary build, install script, CI/CD pipeline, dan dokumentasi  
> **Prasyarat:** Phase 03 selesai. Semua fitur inti berjalan stabil dan seluruh test suite pass  
> **Output yang diharapkan:** Symphony dapat diinstal oleh siapapun dengan satu perintah curl, binary tersedia untuk semua platform utama, dan ekosistem plugin terbuka untuk kontribusi eksternal

---

## Konteks untuk AI

Fase ini adalah transisi Symphony dari "tool pribadi yang berfungsi" ke "tool yang siap dipakai publik." Tidak ada fitur engine baru di fase ini — semua pekerjaan bersifat infrastruktur, tooling, dan dokumentasi.

Tiga hal yang paling menentukan kesan pertama komunitas open source adalah: kemudahan instalasi, keandalan build lintas platform, dan kualitas dokumentasi awal. Ketiga hal ini adalah fokus utama fase ini.

---

## Tugas 1 — Plugin System (`internal/engine/`)

Implementasikan plugin system yang memungkinkan pengguna mendefinisikan custom renderer untuk tipe file yang tidak didukung secara native oleh `text/template`.

### Arsitektur Plugin: Executable-Based

Plugin adalah binary eksternal yang berkomunikasi dengan Symphony via stdin/stdout menggunakan protokol JSON. Ini menghindari kompleksitas CGO atau dynamic linking, dan memungkinkan plugin ditulis dalam bahasa apapun.

```
Symphony Engine
      │
      │  stdin → {"context": {...}, "file_content": "..."}
      ▼
  Plugin Binary (Go/Python/Node/Rust — bebas)
      │
      │  stdout → {"rendered_content": "..."}
      ▼
Symphony Engine
```

### Interface Plugin (`internal/engine/plugin.go`)

```go
// PluginRenderer menjalankan plugin eksternal untuk me-render file
type PluginRenderer struct {
    Name       string
    Executable string
    Handles    []string  // Glob patterns, e.g. ["*.prisma", "*.proto"]
}

// PluginRequest adalah payload yang dikirim ke plugin via stdin
type PluginRequest struct {
    Context     map[string]any `json:"context"`
    FileContent string         `json:"file_content"`
    SourcePath  string         `json:"source_path"`
    TargetPath  string         `json:"target_path"`
}

// PluginResponse adalah payload yang diterima dari plugin via stdout
type PluginResponse struct {
    RenderedContent string `json:"rendered_content"`
    Error           string `json:"error,omitempty"`
}

// Render menjalankan plugin binary dan mengembalikan konten yang sudah di-render
func (p *PluginRenderer) Render(ctx *EngineContext, sourceContent string) (string, error)
```

### Ketentuan Keamanan Plugin

Sebelum menjalankan plugin, validasi bahwa executable:
1. Ada di path yang ditentukan
2. Memiliki permission yang dapat dieksekusi
3. Bukan path yang mengandung `..` (path traversal prevention)

Timeout eksekusi plugin: 30 detik. Jika plugin tidak merespons dalam 30 detik, batalkan dan kembalikan error.

### Registrasi Plugin di Engine

Di dalam `engine.go`, sebelum memproses setiap file, periksa apakah ada plugin yang terdaftar yang bisa menangani file tersebut (berdasarkan glob pattern di `handles`). Jika ada, delegasikan ke plugin alih-alih `text/template` renderer.

---

## Tugas 2 — Finalisasi `.goreleaser.yaml`

Update konfigurasi goreleaser yang dibuat sebagai kerangka di Fase 0 menjadi konfigurasi yang production-ready.

```yaml
# .goreleaser.yaml
version: 2

project_name: symphony

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: symphony
    main: ./main.go
    binary: symphony
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/username/symphony/cmd.Version={{.Version}}
      - -X github.com/username/symphony/cmd.Commit={{.Commit}}
      - -X github.com/username/symphony/cmd.BuildDate={{.Date}}

archives:
  - id: symphony
    builds: [symphony]
    name_template: "symphony_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE
      - docs/*

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

signs:
  - artifacts: checksum
    args: ["--batch", "-u", "{{ .Env.GPG_FINGERPRINT }}", "--output", "${signature}", "--detach-sign", "${artifact}"]

release:
  github:
    owner: username
    name: symphony
  draft: false
  prerelease: auto
  name_template: "Symphony v{{.Version}}"
  footer: |
    ## Installation
    ```bash
    curl -sSL https://raw.githubusercontent.com/username/symphony/main/install.sh | sh
    ```

changelog:
  sort: asc
  use: github
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - "Merge pull request"
  groups:
    - title: "🚀 New Features"
      regexp: "^feat"
      order: 0
    - title: "🐛 Bug Fixes"
      regexp: "^fix"
      order: 1
    - title: "⚡ Performance"
      regexp: "^perf"
      order: 2

brews:
  - name: symphony
    homepage: "https://github.com/username/symphony"
    description: "The Adaptive Scaffolding Engine"
    tap:
      owner: username
      name: homebrew-symphony
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
    install: |
      bin.install "symphony"
    test: |
      system "#{bin}/symphony version"
```

---

## Tugas 3 — Install Script

Buat shell script yang memungkinkan instalasi Symphony dengan satu perintah.

### `install.sh`

Script harus:
1. Deteksi OS dan arsitektur secara otomatis (Linux/macOS/Windows via WSL, amd64/arm64)
2. Unduh binary yang sesuai dari GitHub Releases
3. Verifikasi checksum SHA256
4. Pindahkan binary ke direktori yang ada di `$PATH` (default: `/usr/local/bin` di Unix, `%LOCALAPPDATA%\Programs` di Windows)
5. Verifikasi instalasi dengan menjalankan `symphony version`
6. Tampilkan pesan sukses atau petunjuk troubleshooting jika gagal

```bash
#!/bin/sh
# Symphony CLI Installer
# Usage: curl -sSL https://github.com/username/symphony/releases/latest/download/install.sh | sh

set -e

REPO="username/symphony"
BINARY_NAME="symphony"

# Deteksi OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Arsitektur tidak didukung: $ARCH"; exit 1 ;;
esac

# [Lanjutkan dengan download, checksum verification, dan instalasi]
```

---

## Tugas 4 — GitHub Actions CI/CD Pipeline

Buat dua workflow GitHub Actions.

### `.github/workflows/ci.yml` — Continuous Integration

Dijalankan pada setiap push ke semua branch dan semua pull request.

```yaml
name: CI

on:
  push:
    branches: ["*"]
  pull_request:
    branches: [main]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ["1.22", "1.23"]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      - name: Download dependencies
        run: go mod download
      - name: Run linter
        uses: golangci/golangci-lint-action@v4
      - name: Run tests
        run: go test ./... -race -coverprofile=coverage.out
      - name: Upload coverage
        uses: codecov/codecov-action@v4
```

### `.github/workflows/release.yml` — Release Pipeline

Dijalankan hanya ketika tag baru dengan format `v*` di-push ke branch main.

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ secrets.GPG_FINGERPRINT }}
```

---

## Tugas 5 — Version Information di Binary

Inject informasi build ke dalam binary saat kompilasi menggunakan ldflags.

```go
// cmd/version.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

// Variabel ini di-inject saat build time via ldflags
var (
    Version   = "dev"
    Commit    = "none"
    BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Tampilkan versi Symphony",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Symphony CLI\n")
        fmt.Printf("  Version   : %s\n", Version)
        fmt.Printf("  Commit    : %s\n", Commit)
        fmt.Printf("  Built     : %s\n", BuildDate)
    },
}
```

Tambahkan juga logika perbandingan versi: saat `--verbose` aktif, Symphony memeriksa versi terbaru di GitHub Releases API secara asinkronus dan menampilkan notifikasi jika ada update — tanpa memblokir eksekusi command utama.

---

## Tugas 6 — Dokumentasi Publik

### `README.md`

Buat README yang komprehensif namun ringkas. README harus mencakup:

Bagian pertama adalah quick demo — GIF atau ASCII art yang menunjukkan Symphony bekerja dalam 30 detik. Ini adalah hal pertama yang dilihat orang.

Bagian kedua adalah installation instructions — satu perintah curl untuk Unix, instruksi manual untuk Windows.

Bagian ketiga adalah quick start — contoh paling sederhana dari awal sampai akhir menggunakan template lokal sederhana yang ikut disertakan di repositori.

Bagian keempat adalah template YAML reference — dokumentasi singkat semua field yang tersedia di `template.yaml` dengan contoh.

Bagian kelima adalah contributing guide — bagaimana cara membuat template sendiri dan cara berkontribusi ke Symphony core.

### `docs/blueprint-spec.md`

Dokumentasi lengkap dan formal untuk setiap field dalam `template.yaml`. Ini adalah referensi untuk pembuat template. Setiap field harus memiliki: tipe data, apakah wajib, nilai default, dan minimal satu contoh penggunaan.

### `docs/contributing.md`

Panduan untuk kontributor yang ingin menambahkan fitur ke Symphony core. Mencakup: cara setup development environment, cara menjalankan test suite, konvensi commit message (Conventional Commits), dan proses pull request review.

---

## Tugas 7 — `golangci-lint` Configuration

Buat `.golangci.yml` di root proyek untuk konfigurasi linter yang konsisten.

```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gosec         # Security checks
    - misspell
    - gocyclo       # Cyclomatic complexity
    - dupl          # Duplicate code detection

linters-settings:
  gocyclo:
    min-complexity: 15
  gosec:
    excludes:
      - G304  # File path provided as taint input — kita memang perlu ini untuk template fetching

issues:
  exclude-rules:
    - path: "_test.go"
      linters: [gosec, dupl]
```

---

## Tugas 8 — `Makefile` Final

Update Makefile dengan semua target yang diperlukan untuk workflow development dan release.

```makefile
.PHONY: all build test lint clean release install snapshot

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags="-X 'github.com/username/symphony/cmd.Version=$(VERSION)' \
                     -X 'github.com/username/symphony/cmd.Commit=$(COMMIT)' \
                     -X 'github.com/username/symphony/cmd.BuildDate=$(DATE)'"

all: lint test build

build:
	@go build $(LDFLAGS) -o ./bin/symphony ./main.go

test:
	@go test ./... -race -v

test-cover:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@golangci-lint run ./...

clean:
	@rm -rf ./bin coverage.out coverage.html dist/

# Build snapshot untuk testing (tanpa publish ke GitHub)
snapshot:
	@goreleaser release --snapshot --clean

# Build release resmi (membutuhkan tag dan GitHub token)
release:
	@goreleaser release --clean

install:
	@go install $(LDFLAGS) ./main.go

tidy:
	@go mod tidy
```

---

## Checklist Selesai Fase 4

- [ ] `goreleaser release --snapshot` menghasilkan binary untuk Linux, macOS (Intel + ARM), dan Windows tanpa error
- [ ] Binary yang dihasilkan bisa dijalankan di masing-masing platform
- [ ] `symphony version` menampilkan versi, commit, dan tanggal build dengan benar
- [ ] Install script berhasil menginstal Symphony di Linux dan macOS
- [ ] GitHub Actions CI pipeline berjalan hijau di semua kombinasi OS dan Go version
- [ ] Plugin system: Symphony bisa memanggil plugin binary eksternal dan menggunakan output-nya
- [ ] `README.md` mencakup instalasi, quick start, dan referensi singkat
- [ ] `docs/blueprint-spec.md` mendokumentasikan semua field `template.yaml`
- [ ] `golangci-lint run ./...` tidak menghasilkan error atau warning

---

## Catatan Penting untuk AI

Untuk install script, selalu prioritaskan keamanan: verifikasi checksum sebelum menjalankan binary, jangan pernah pipe langsung ke `sh` tanpa verifikasi (meskipun itu yang ditampilkan di dokumentasi sebagai convenience). Tambahkan flag `--verify-only` di install script untuk pengguna yang ingin memverifikasi download tanpa langsung menginstal.

Untuk dokumentasi, tulis dalam bahasa Inggris agar dapat dijangkau oleh komunitas internasional sejak awal. Dokumentasi yang baik akan menentukan apakah orang akan mencoba tool ini setelah menemukannya di GitHub.
