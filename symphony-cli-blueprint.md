# Symphony CLI — Enterprise Blueprint
> *The Adaptive Scaffolding Engine*

**Version:** 0.1.0-draft  
**Status:** Pre-Development · Personal Use → Open Source  
**Last Updated:** 2025-03-25  

---

## Table of Contents

1. [Vision & Philosophy](#1-vision--philosophy)
2. [Tech Stack](#2-tech-stack)
3. [Project Structure](#3-project-structure)
4. [Terminal UI/UX Design](#4-terminal-uiux-design)
5. [Architecture Design](#5-architecture-design)
6. [Template & Blueprint System](#6-template--blueprint-system)
7. [Core Features](#7-core-features)
8. [CLI Commands & UX Flow](#8-cli-commands--ux-flow)
9. [Workflow Engine](#9-workflow-engine)
10. [AST Injection System](#10-ast-injection-system)
11. [Remote Template Fetching](#11-remote-template-fetching)
12. [Lock File & Reproducibility](#12-lock-file--reproducibility)
13. [Plugin System](#13-plugin-system)
14. [Error Handling Strategy](#14-error-handling-strategy)
15. [Testing Strategy](#15-testing-strategy)
16. [Development Roadmap](#16-development-roadmap)
17. [Architect's Notes & Warnings](#17-architects-notes--warnings)

---

## 1. Vision & Philosophy

Symphony bukan sekadar code generator. Symphony adalah sebuah **orchestrator** yang memahami hubungan antar komponen dalam arsitektur software — membaca konteks dari jawaban user, menyesuaikan output secara dinamis, dan mengotomasi ritual-ritual repetitif yang menghabiskan waktu seorang developer.

### Core Principles

| Prinsip | Penjelasan |
|---|---|
| **Data-Driven Core** | Engine dan Template sepenuhnya terpisah. Logika CLI tidak mengetahui isi template sama sekali. |
| **Zero Assumption** | CLI tidak berasumsi tentang stack yang digunakan. Semua keputusan didorong oleh file `template.yaml`. |
| **Reproducibility First** | Setiap hasil scaffolding harus bisa direproduksi secara identik kapanpun dan di mesin manapun. |
| **Fail Loudly** | Validasi dilakukan sedini mungkin. Kesalahan input tidak boleh baru terdeteksi setelah proses generasi selesai. |
| **Composable by Design** | Template bisa mewarisi dan meng-compose template lain, bukan menduplikasi. |

### Positioning

```
Cookiecutter  ──────────────────────────────── Nx / Turbo
(Terlalu sederhana,                   (Terlalu opinionated,
 tidak ada kondisional)                 hanya untuk Node.js)

                      [ Symphony ]
                   Multi-language ✓
                   Composable     ✓
                   Single binary  ✓
                   Conditional    ✓
```

---

## 2. Tech Stack

### 2.1 Bahasa Utama

| Komponen | Teknologi | Alasan |
|---|---|---|
| **CLI Runtime** | Go 1.22+ | Single binary distribution, zero-dependency, startup < 50ms |
| **Template Engine** | `text/template` (stdlib) | Built-in Go, powerful, zero external dependency |
| **Config Parsing** | YAML via `gopkg.in/yaml.v3` | Human-readable, mendukung anchors & multi-document |
| **Schema Validation** | `github.com/xeipuuv/gojsonschema` | Validasi `template.yaml` sebelum eksekusi |

### 2.2 CLI Framework & Command Layer

| Library | Versi | Fungsi |
|---|---|---|
| `cobra` | v1.8+ | Command structure (`gen`, `check`, `list`, `update`) — standar industri (Kubernetes, Docker) |
| `viper` | v1.18+ | Manajemen konfigurasi user global (`~/.symphony/config.yaml`) |

### 2.3 Terminal UI/UX Layer

| Library | Versi | Fungsi |
|---|---|---|
| `bubbletea` | v0.26+ | Framework TUI berbasis Elm Architecture — untuk interactive prompts dan layout |
| `bubbles` | v0.18+ | Komponen siap pakai: `textinput`, `list`, `progress`, `spinner`, `table` |
| `lipgloss` | v0.10+ | Styling terminal: warna, border, padding, alignment |
| `glamour` | v0.7+ | Render Markdown di terminal — untuk menampilkan instruksi post-scaffold |

### 2.4 Networking & Remote Fetching

| Library | Fungsi |
|---|---|
| `go-getter` (HashiCorp) | Download template dari GitHub, GitLab, S3, dan URL arbitrer |
| `go-github` | GitHub API client — untuk validasi template version dan tag |
| `archiver` | Extract `.zip` dan `.tar.gz` dari remote template |

### 2.5 AST & Code Intelligence

| Library | Target Language | Fungsi |
|---|---|---|
| `go/ast` + `go/parser` (stdlib) | Go | Parse dan inject ke file `.go` |
| `tree-sitter-go` (via CGO binding) | Go (advanced) | Analisis semantik yang lebih dalam |
| Regex-based injector | TypeScript, Python, dll | Pendekatan pragmatis untuk bahasa non-Go |

### 2.6 Tooling Tambahan

| Tool | Fungsi |
|---|---|
| `goreleaser` | Build dan release binary lintas platform (Linux, macOS, Windows) |
| `cobra-cli` | Generator untuk command baru selama development |
| `golangci-lint` | Linting dan code quality |
| `testify` | Unit testing assertions |

---

## 3. Project Structure

```
symphony/
│
├── cmd/                          # Entry points untuk setiap command Cobra
│   ├── root.go                   # Root command + global flags
│   ├── gen.go                    # `symphony gen` command
│   ├── check.go                  # `symphony check` command
│   ├── list.go                   # `symphony list` command
│   └── update.go                 # `symphony update` command
│
├── internal/
│   ├── engine/                   # Core scaffolding engine
│   │   ├── engine.go             # Orchestrator utama
│   │   ├── walker.go             # File tree walker
│   │   ├── renderer.go           # text/template renderer
│   │   ├── writer.go             # File writer dengan dry-run support
│   │   └── hooks.go              # Post-scaffold hook runner
│   │
│   ├── blueprint/                # Parser & validator untuk template.yaml
│   │   ├── parser.go
│   │   ├── validator.go
│   │   ├── schema.go             # JSON Schema definition
│   │   └── resolver.go           # Dependency graph resolver untuk prompts
│   │
│   ├── ast/                      # AST injection module
│   │   ├── go_injector.go
│   │   ├── ts_injector.go
│   │   └── injector_interface.go
│   │
│   ├── remote/                   # Remote template fetching
│   │   ├── fetcher.go
│   │   ├── github.go
│   │   └── cache.go              # Local cache di ~/.symphony/cache/
│   │
│   ├── tui/                      # Terminal UI components (Bubbletea)
│   │   ├── prompt.go             # Interactive prompt orchestrator
│   │   ├── select.go             # Single-select component
│   │   ├── multiselect.go        # Multi-select component
│   │   ├── input.go              # Text input component
│   │   ├── confirm.go            # Yes/No confirmation
│   │   ├── progress.go           # Progress bar
│   │   └── summary.go            # Pre-execution diff summary
│   │
│   ├── lock/                     # Lock file management
│   │   ├── writer.go
│   │   └── reader.go
│   │
│   └── config/                   # Global user config (~/.symphony/)
│       └── config.go
│
├── pkg/
│   └── expr/                     # Expression evaluator untuk kondisi `if:`
│       └── eval.go               # Parser untuk "DB_TYPE != 'none'" dsb
│
├── testdata/                     # Template fixtures untuk testing
│   └── templates/
│       ├── simple-go/
│       └── hexagonal-go/
│
├── docs/                         # Dokumentasi
│   ├── blueprint-spec.md         # Spesifikasi lengkap template.yaml
│   └── contributing.md
│
├── main.go
├── go.mod
├── go.sum
├── .goreleaser.yaml
└── Makefile
```

---

## 4. Terminal UI/UX Design

### 4.1 Design Principles

Antarmuka terminal Symphony mengikuti prinsip **Progressive Disclosure** — informasi ditampilkan hanya saat relevan, tidak membanjiri user di awal.

```
Clarity      > Aesthetic
Feedback     > Silence
Confirmation > Assumption
```

### 4.2 Screen Flow

```
┌─────────────────────────────────────────────────────────┐
│                    SYMPHONY CLI v0.4.0                  │
│              The Adaptive Scaffolding Engine            │
└─────────────────────────────────────────────────────────┘

         [1. Discovery]  →  [2. Prompting]  →  [3. Preview]  →  [4. Generation]  →  [5. Summary]
```

### 4.3 Tampilan Setiap Fase

#### Fase 1 — Discovery Screen
```
  ◆ Symphony                                    v0.4.0
  ─────────────────────────────────────────────────────
  ✔ Template ditemukan: github.com/user/go-blueprint
  ✔ Versi: 1.2.0 (commit: a3f8c12)
  ✔ Kompatibel dengan Symphony >= 0.3.0
  
  Loading blueprint...  ████████████████████  100%
```

#### Fase 2 — Interactive Prompting
```
  ◆ Konfigurasi Proyek Baru
  ─────────────────────────────────────────────────────

  ? Nama proyek:
  ❯ my-awesome-api
    ↑↓ ketik nama proyekmu

  ? Database apa yang akan digunakan?
  ❯ ● PostgreSQL
    ○ MongoDB
    ○ MySQL
    ○ None

  ? Gunakan Redis untuk caching? (y/N)
```

#### Fase 3 — Pre-Execution Diff Preview
```
  ◆ Review — File yang akan dibuat:
  ─────────────────────────────────────────────────────
  
  [CREATE]  cmd/main.go
  [CREATE]  internal/domain/user/entity.go
  [CREATE]  internal/usecase/user/service.go
  [CREATE]  internal/infrastructure/postgres/db.go
  [MODIFY]  go.mod                              (+3 deps)
  [SKIP]    internal/infrastructure/redis/      (Redis dinonaktifkan)

  Total: 12 file akan dibuat, 1 file dimodifikasi

  ❯ Lanjutkan? (Y/n)
```

#### Fase 4 — Generation Progress
```
  ◆ Generating...
  ─────────────────────────────────────────────────────
  
  ✔ cmd/main.go
  ✔ internal/domain/user/entity.go
  ✔ internal/usecase/user/service.go
  ⠸ internal/infrastructure/postgres/db.go

  Progress  ████████████░░░░░░░░  8/12 files
```

#### Fase 5 — Completion Summary
```
  ◆ Scaffold Selesai!
  ─────────────────────────────────────────────────────
  
  ✔ 12 file dibuat
  ✔ go mod tidy dijalankan
  ✔ git init dijalankan
  ✔ symphony.lock dibuat

  ─────────────────────────────────────────────────────
  
  ▶  Langkah Selanjutnya:
  
     cd my-awesome-api
     cp .env.example .env
     docker-compose up -d
     go run ./cmd/main.go
  
  ─────────────────────────────────────────────────────
  ✨ Happy coding!
```

### 4.4 Warna & Styling (Lipgloss Palette)

| Elemen | Warna | Hex |
|---|---|---|
| **Header / Brand** | Cyan Bold | `#00D7FF` |
| **Success** `✔` | Green | `#04B575` |
| **Warning** `⚠` | Yellow | `#FFD700` |
| **Error** `✖` | Red | `#FF5F57` |
| **Dimmed / Secondary** | Gray | `#626262` |
| **Highlight / Selection** | Magenta | `#D787FF` |
| **Progress Bar Fill** | Blue | `#5F87FF` |
| **File: CREATE** | Green | `#04B575` |
| **File: MODIFY** | Yellow | `#FFD700` |
| **File: SKIP** | Gray | `#626262` |
| **File: DELETE** | Red | `#FF5F57` |

### 4.5 Keyboard Shortcuts

| Key | Aksi |
|---|---|
| `↑` / `↓` | Navigasi pilihan |
| `Space` | Toggle pilihan (multi-select) |
| `Enter` | Konfirmasi |
| `Ctrl+C` | Batalkan dan keluar dengan bersih |
| `?` | Tampilkan bantuan inline |
| `Tab` | Autocomplete (text input) |

---

## 5. Architecture Design

### 5.1 High-Level Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                      │
│              gen │ check │ list │ update │ re-gen             │
└───────────────────────────────┬───────────────────────────────┘
                                │
┌───────────────────────────────▼───────────────────────────────┐
│                     Orchestrator (Engine)                     │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  Blueprint  │  │   Context   │  │    Action Runner    │   │
│  │   Parser    │→ │   Builder   │→ │  (render/shell/ast) │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└────────────┬──────────────────────────────────────┬──────────┘
             │                                      │
┌────────────▼────────────┐          ┌──────────────▼──────────┐
│    Template Layer       │          │    Output Layer         │
│  (YAML + .tmpl files)   │          │  (File Writer + Hooks)  │
│  Local or Remote        │          │  + Lock File Generator  │
└─────────────────────────┘          └─────────────────────────┘
```

### 5.2 Context Object

Seluruh jawaban user dikumpulkan ke dalam satu `Context` map yang mengalir ke semua bagian engine:

```go
// internal/engine/context.go
type Context struct {
    Values    map[string]interface{}  // Jawaban dari semua prompt
    Meta      TemplateMeta            // Info tentang template yang digunakan
    OutputDir string
    DryRun    bool
}
```

### 5.3 Expression Evaluator

Kondisi `if:` di dalam `actions` dievaluasi oleh expression evaluator khusus yang ringan — bukan `eval()` atau `exec()`. Ini penting untuk keamanan.

```
"DB_TYPE != 'none' && USE_REDIS == true"
        │
        ▼
  Tokenizer → Parser → AST → Evaluator(context.Values)
        │
        ▼
      bool
```

---

## 6. Template & Blueprint System

### 6.1 Struktur Folder Template

```
my-go-template/
├── template.yaml            # Blueprint definition (otak dari template)
├── README.md.tmpl           # Template file — akan di-render
├── cmd/
│   └── main.go.tmpl
├── internal/
│   ├── domain/
│   │   └── entity.go.tmpl
│   └── infrastructure/
│       ├── postgres/
│       │   └── db.go.tmpl   # Hanya dibuat jika DB_TYPE == PostgreSQL
│       └── redis/
│           └── cache.go.tmpl # Hanya dibuat jika USE_REDIS == true
└── .hooks/
    ├── post-scaffold.sh     # Dijalankan setelah semua file selesai dibuat
    └── pre-scaffold.sh      # Validasi environment sebelum mulai
```

### 6.2 Spesifikasi `template.yaml`

```yaml
# ─── Metadata ────────────────────────────────────────────────
schema_version: "2"
name: "Clean-Hexagonal-Go"
version: "1.2.0"
author: "username"
description: "Production-ready Go REST API dengan hexagonal architecture"
min_symphony_version: "0.3.0"
tags: ["go", "hexagonal", "postgresql", "rest-api"]

# ─── Inheritance / Composition ───────────────────────────────
extends: "github.com/user/symphony-base/go-core@v1.0.0"  # Opsional

# ─── Input Validation ────────────────────────────────────────
validations:
  - field: "PROJECT_NAME"
    rule: "regex"
    pattern: "^[a-z][a-z0-9-]+$"
    message: "Nama proyek harus lowercase, tidak boleh mengandung spasi."
  - field: "MODULE_PATH"
    rule: "required"

# ─── Prompts ─────────────────────────────────────────────────
prompts:
  - id: "PROJECT_NAME"
    question: "Nama proyek:"
    type: "input"
    default: "my-go-api"

  - id: "MODULE_PATH"
    question: "Go module path:"
    type: "input"
    default: "github.com/username/{{.PROJECT_NAME}}"

  - id: "DB_TYPE"
    question: "Database yang akan digunakan?"
    type: "select"
    options: ["PostgreSQL", "MongoDB", "MySQL", "None"]
    default: "PostgreSQL"

  - id: "DB_HOST"
    question: "Hostname database:"
    type: "input"
    default: "localhost"
    depends_on: "DB_TYPE != 'None'"   # Hanya muncul jika DB dipilih

  - id: "USE_MIGRATIONS"
    question: "Gunakan database migration tool (golang-migrate)?"
    type: "confirm"
    default: true
    depends_on: "DB_TYPE == 'PostgreSQL'"  # Hanya relevan untuk PostgreSQL

  - id: "USE_REDIS"
    question: "Aktifkan Redis untuk caching?"
    type: "confirm"
    default: false

  - id: "USE_TRACING"
    question: "Aktifkan distributed tracing (OpenTelemetry)?"
    type: "confirm"
    default: false

# ─── Actions ─────────────────────────────────────────────────
actions:
  # Render file dasar (selalu dieksekusi)
  - type: "render"
    source: "./cmd/main.go.tmpl"
    target: "./cmd/main.go"

  # Render file database (kondisional)
  - type: "render"
    source: "./internal/infrastructure/postgres/db.go.tmpl"
    target: "./internal/infrastructure/postgres/db.go"
    if: "DB_TYPE == 'PostgreSQL'"

  - type: "render"
    source: "./internal/infrastructure/redis/cache.go.tmpl"
    target: "./internal/infrastructure/redis/cache.go"
    if: "USE_REDIS == true"

  # AST Injection — daftarkan route baru ke main.go
  - type: "ast-inject"
    target: "./cmd/main.go"
    strategy: "after"
    anchor: "// ROUTES_INJECTION_POINT"
    content: "r.Handle(\"/api/v1\", apiRouter)"
    if: "ENABLE_API_V1 == true"

  # Post-scaffold hooks
  - type: "shell"
    command: "go mod tidy"
    working_dir: "{{.OUTPUT_DIR}}"

  - type: "shell"
    command: "git init && git add . && git commit -m 'chore: initial scaffold by Symphony'"
    working_dir: "{{.OUTPUT_DIR}}"

# ─── Post-Scaffold Message ───────────────────────────────────
completion_message: |
  ## Proyek Berhasil Dibuat!

  Langkah selanjutnya:

  ```bash
  cd {{.PROJECT_NAME}}
  cp .env.example .env
  docker-compose up -d
  go run ./cmd/main.go
  ```
```

### 6.3 Template File (`.tmpl` Syntax)

```go
// cmd/main.go.tmpl
package main

import (
    "log"
    "{{.MODULE_PATH}}/internal/infrastructure/server"
    {{- if eq .DB_TYPE "PostgreSQL"}}
    "{{.MODULE_PATH}}/internal/infrastructure/postgres"
    {{- end}}
    {{- if .USE_REDIS}}
    "{{.MODULE_PATH}}/internal/infrastructure/redis"
    {{- end}}
)

func main() {
    {{- if eq .DB_TYPE "PostgreSQL"}}
    db, err := postgres.NewConnection()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    {{- end}}

    // ROUTES_INJECTION_POINT

    srv := server.New(db)
    log.Fatal(srv.Start(":8080"))
}
```

---

## 7. Core Features

### 7.1 Dynamic Blueprinting
Seluruh alur tanya-jawab dan file generation didefinisikan sepenuhnya dalam `template.yaml`. Engine tidak memiliki hardcoded logic untuk bahasa atau framework apapun.

### 7.2 Conditional Scaffolding
Setiap action mendukung ekspresi kondisional pada field `if:`. Folder dan file yang tidak relevan tidak akan dibuat — bukan sekadar dibuat kosong.

### 7.3 Dry-Run Mode
Sebelum menulis ke disk, Symphony menampilkan preview lengkap: file apa yang akan dibuat, dimodifikasi, atau dilewati. User harus memberikan konfirmasi eksplisit sebelum eksekusi dimulai.

### 7.4 Dependency-Aware Prompts
Pertanyaan muncul secara dinamis berdasarkan jawaban sebelumnya. Jika user memilih "None" untuk database, semua pertanyaan terkait database otomatis dilewati.

### 7.5 Template Composition (Inheritance)
Template dapat meng-`extend` template lain. Child template hanya perlu mendefinisikan delta — prompt tambahan dan action tambahan — tanpa menduplikasi boilerplate dari base template.

### 7.6 AST Injection
Kemampuan untuk menyisipkan baris kode ke file yang sudah ada secara terprogram — mendaftarkan route baru, menambahkan import, atau menyisipkan dependency injection — tanpa merusak struktur file.

### 7.7 Remote Template Fetching & Caching
Template dapat diunduh langsung dari GitHub, GitLab, atau URL publik. Template yang diunduh di-cache secara lokal di `~/.symphony/cache/` untuk penggunaan offline.

### 7.8 Lock File & Re-generation
Setiap scaffold menghasilkan `symphony.lock` yang merekam semua input dan versi template yang digunakan. Command `symphony re-gen` dapat mereproduksi scaffold yang identik dari lock file ini.

### 7.9 Post-Scaffold Hooks
Setelah semua file dibuat, Symphony menjalankan perintah shell yang didefinisikan dalam template — seperti `go mod tidy`, `git init`, atau `npm install`.

### 7.10 Template Health Check
Command `symphony check` memvalidasi bahwa template yang digunakan masih kompatibel: memeriksa versi minimum Symphony, memvalidasi schema, dan (opsional) mengecek apakah dependency yang diinjeksi ke `go.mod` masih valid di pkg.go.dev.

---

## 8. CLI Commands & UX Flow

### 8.1 Daftar Commands

```
symphony
├── gen       <source> [flags]    # Scaffold proyek baru
├── re-gen    [flags]             # Ulangi scaffold dari lock file
├── check     <source>           # Validasi template tanpa menjalankannya
├── list                         # Tampilkan template yang tersedia / ter-cache
├── update    <source>           # Update template ke versi terbaru
├── cache
│   ├── list                     # Tampilkan cache lokal
│   └── clear                    # Hapus cache lokal
└── version                      # Tampilkan versi Symphony CLI
```

### 8.2 Global Flags

| Flag | Shorthand | Deskripsi |
|---|---|---|
| `--out <dir>` | `-o` | Direktori output (default: direktori saat ini) |
| `--dry-run` | `-d` | Preview tanpa menulis ke disk |
| `--no-hooks` | | Lewati post-scaffold hooks |
| `--yes` | `-y` | Skip semua konfirmasi (untuk CI/CD) |
| `--verbose` | `-v` | Tampilkan log detail |
| `--config <file>` | `-c` | Gunakan config file kustom |

### 8.3 Contoh Penggunaan

```bash
# Instalasi
curl -sSL https://symphony.dev/install.sh | sh

# Scaffold dari template lokal
symphony gen ./templates/go-hexagonal --out ./my-new-api

# Scaffold dari GitHub (versi spesifik)
symphony gen github.com/user/go-blueprint@v1.2.0 --out ./my-new-api

# Preview tanpa menulis ke disk
symphony gen github.com/user/go-blueprint --out ./my-api --dry-run

# Re-generate dari lock file (di dalam direktori proyek)
symphony re-gen

# Validasi template tanpa menjalankan scaffold
symphony check github.com/user/go-blueprint

# Scaffold tanpa konfirmasi (untuk CI/CD pipeline)
symphony gen ./templates/go-api --out ./app --yes
```

---

## 9. Workflow Engine

### 9.1 Tahap 1 — Discovery & Validation

```
symphony gen <source>
      │
      ▼
  Resolve source (local path / GitHub URL / cache)
      │
      ▼
  Load template.yaml
      │
      ▼
  Validate schema_version & min_symphony_version
      │
      ▼
  Resolve inheritance (jika ada `extends:`)
      │
      ▼
  Merge base + child blueprint
```

### 9.2 Tahap 2 — Interactive Prompting

```
  Build prompt dependency graph (topological sort)
      │
      ▼
  Tampilkan prompt satu per satu (Bubbletea TUI)
      │
      ├── Prompt dengan depends_on → evaluasi kondisi
      │   └── Jika false → skip prompt
      │
      ▼
  Jalankan validasi per-field (regex, required, dll)
      │
      ▼
  Build Context object dari semua jawaban
```

### 9.3 Tahap 3 — Dry-Run Preview

```
  Walk semua file di folder template
      │
      ▼
  Untuk setiap action:
      ├── Evaluasi ekspresi `if:` dengan Context
      ├── Jika false → tandai sebagai [SKIP]
      └── Jika true  → tandai sebagai [CREATE] / [MODIFY]
      │
      ▼
  Tampilkan diff summary di terminal
      │
      ▼
  Tunggu konfirmasi user
```

### 9.4 Tahap 4 — File Generation

```
  Untuk setiap action yang lolos filter:
      │
      ├── type: render   → text/template + Context → tulis file
      ├── type: ast-inject → baca file target, inject, tulis ulang
      └── type: shell    → exec.Command dengan working_dir
      │
      ▼
  Tampilkan progress bar real-time
```

### 9.5 Tahap 5 — Finalization

```
  Jalankan semua post-scaffold hooks secara berurutan
      │
      ▼
  Generate symphony.lock
      │
      ▼
  Render completion_message (via Glamour Markdown renderer)
      │
      ▼
  Exit 0
```

---

## 10. AST Injection System

### 10.1 Strategi Injeksi

Symphony menggunakan dua pendekatan injeksi berdasarkan kebutuhan:

| Pendekatan | Kapan Digunakan | Keunggulan |
|---|---|---|
| **Regex-based** (Fase 1) | Injeksi sederhana ke semua bahasa | Mudah diimplementasikan, tidak ada dependency eksternal |
| **`go/ast` Parser** (Fase 3) | Injeksi semantik ke file `.go` | Presisi tinggi, tidak bisa salah posisi |

### 10.2 Konfigurasi Injeksi di `template.yaml`

```yaml
# Strategi 1: Inject via anchor comment (regex-based)
- type: "ast-inject"
  target: "./cmd/main.go"
  strategy: "after-anchor"
  anchor: "// ROUTES_INJECTION_POINT"
  content: |
    r.Group("/auth", authHandler.RegisterRoutes)

# Strategi 2: Inject setelah function declaration (go/ast)
- type: "ast-inject"
  target: "./cmd/main.go"
  strategy: "after-func"
  anchor: "func setupRoutes"
  content: |
    r.Handle("/metrics", promhttp.Handler())
```

### 10.3 Prinsip Keamanan AST

Sebelum melakukan injeksi, engine selalu:
1. Memverifikasi bahwa file target ada dan bisa di-parse.
2. Membuat backup file target (`file.go.symphony-bak`).
3. Menjalankan `gofmt` pada hasil injeksi sebelum menulis.
4. Jika `gofmt` gagal, membatalkan injeksi dan mengembalikan backup.

---

## 11. Remote Template Fetching

### 11.1 Source Format yang Didukung

```bash
# GitHub (tag atau branch spesifik)
symphony gen github.com/user/repo@v1.2.0

# GitHub (branch terbaru)
symphony gen github.com/user/repo

# GitLab
symphony gen gitlab.com/user/repo@main

# URL langsung (zip atau tar.gz)
symphony gen https://example.com/template.tar.gz

# Template lokal
symphony gen ./my-local-template
```

### 11.2 Caching Strategy

```
~/.symphony/
├── cache/
│   ├── github.com_user_repo_v1.2.0/    # Immutable (tagged version)
│   └── github.com_user_repo_main/      # Mutable (branch — dicek TTL)
├── config.yaml
└── logs/
```

Aturan cache:
- Template dengan **tag versi** (`@v1.2.0`) di-cache secara permanen karena bersifat immutable.
- Template dengan **branch name** (`@main`) memiliki TTL 24 jam. Setelah TTL, Symphony akan memeriksa apakah ada commit baru sebelum menggunakan cache.

---

## 12. Lock File & Reproducibility

### 12.1 Format `symphony.lock`

```json
{
  "symphony_version": "0.4.1",
  "generated_at": "2025-03-25T10:30:00Z",
  "template": {
    "source": "github.com/user/go-blueprint",
    "version": "1.2.0",
    "commit": "a3f8c12d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b"
  },
  "inputs": {
    "PROJECT_NAME": "my-api",
    "MODULE_PATH": "github.com/user/my-api",
    "DB_TYPE": "PostgreSQL",
    "USE_REDIS": true,
    "USE_TRACING": false
  },
  "output_checksum": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
}
```

### 12.2 Re-generation Flow

```bash
# Di dalam direktori proyek yang sudah di-scaffold
symphony re-gen

# Symphony akan:
# 1. Membaca symphony.lock
# 2. Mengunduh template versi yang sama (dari commit hash)
# 3. Menggunakan inputs yang sama persis
# 4. Menampilkan dry-run preview
# 5. Meminta konfirmasi sebelum overwrite
```

---

## 13. Plugin System

### 13.1 Arsitektur Plugin (Fase 4)

Symphony mendukung custom renderer melalui **executable plugin** yang berkomunikasi via `stdin/stdout` — pola yang sama dengan Terraform providers. Ini menghindari kompleksitas CGO atau dynamic linking.

```
Symphony Engine
      │
      │  [stdin]  JSON: { "context": {...}, "file_content": "..." }
      ▼
  Plugin Binary (go/python/node — bebas)
      │
      │  [stdout] JSON: { "rendered_content": "..." }
      ▼
Symphony Engine
```

### 13.2 Registrasi Plugin di `template.yaml`

```yaml
plugins:
  - name: "custom-renderer"
    executable: "./plugins/my-renderer"
    handles: ["*.prisma", "*.proto"]
```

---

## 14. Error Handling Strategy

### 14.1 Prinsip

Symphony mengikuti prinsip **Fail Early, Explain Clearly**. Setiap error harus memberi tahu user:
1. **Apa** yang salah (pesan singkat).
2. **Di mana** kesalahannya (file atau field yang bermasalah).
3. **Bagaimana** memperbaikinya (saran konkret).

### 14.2 Format Error di Terminal

```
  ✖ Validasi Gagal
  ─────────────────────────────────────────────────────
  
  Field:    PROJECT_NAME
  Nilai:    "My Awesome API"
  Masalah:  Tidak sesuai pola yang diharapkan.
  Pola:     ^[a-z][a-z0-9-]+$
  
  Saran:    Gunakan huruf kecil dan tanda hubung.
            Contoh: "my-awesome-api"

  ─────────────────────────────────────────────────────
  Jalankan ulang: symphony gen <source>
```

### 14.3 Exit Codes

| Code | Kondisi |
|---|---|
| `0` | Sukses |
| `1` | Kesalahan umum (I/O, permission) |
| `2` | Validasi blueprint gagal |
| `3` | Template tidak ditemukan atau tidak kompatibel |
| `4` | User membatalkan (Ctrl+C atau menjawab "No" pada konfirmasi) |
| `5` | Post-scaffold hook gagal |

---

## 15. Testing Strategy

### 15.1 Lapisan Testing

```
Unit Tests          → Fungsi isolasi: parser, evaluator, renderer
Integration Tests   → Engine end-to-end dengan template fixtures
Snapshot Tests      → Output file dicocokkan dengan snapshot yang tersimpan
CLI Tests           → Command-line invocation dengan testdata
```

### 15.2 Contoh Test Case

```go
// internal/engine/renderer_test.go
func TestRenderer_ConditionalBlock(t *testing.T) {
    ctx := Context{Values: map[string]interface{}{
        "DB_TYPE":   "PostgreSQL",
        "USE_REDIS": false,
    }}

    result, err := Render("./testdata/main.go.tmpl", ctx)

    assert.NoError(t, err)
    assert.Contains(t, result, `"postgres"`)          // Postgres block ada
    assert.NotContains(t, result, `"redis"`)          // Redis block tidak ada
}
```

### 15.3 Makefile Targets

```makefile
make test          # Jalankan semua unit tests
make test-int      # Jalankan integration tests
make test-cover    # Generate coverage report
make lint          # Jalankan golangci-lint
make build         # Build binary untuk host OS
make release       # Build semua platform via goreleaser
```

---

## 16. Development Roadmap

### Fase 1 — Core Engine (MVP)
**Goal:** Tool yang bisa dipakai sehari-hari untuk kebutuhan pribadi.

- [ ] Parser `template.yaml` dengan schema validation
- [ ] `text/template` renderer dengan Context injection
- [ ] Command `gen` dasar (tanpa TUI, prompt sederhana)
- [ ] Conditional file generation (`if:` evaluator)
- [ ] Post-scaffold hooks (type `shell`)
- [ ] Generasi `symphony.lock`
- [ ] Unit tests untuk parser, evaluator, renderer

### Fase 2 — UX & Developer Experience
**Goal:** Tool yang nyaman dipakai tanpa perlu membaca dokumentasi.

- [ ] Bubbletea TUI untuk interactive prompts
- [ ] Prompt dependency graph (`depends_on:`)
- [ ] Dry-run preview dengan diff summary
- [ ] Progress bar selama file I/O
- [ ] Command `re-gen` dari lock file
- [ ] Lipgloss styling & warna
- [ ] Glamour Markdown renderer untuk `completion_message`
- [ ] `--yes` flag untuk CI/CD usage

### Fase 3 — Intelligence & Composability
**Goal:** Tool yang bisa digunakan untuk proyek multi-modul yang kompleks.

- [ ] Template inheritance (`extends:`)
- [ ] AST injection (regex-based, Fase 3a)
- [ ] AST injection (go/ast parser, Fase 3b)
- [ ] Command `check` (template health validation)
- [ ] Input validation rules (`regex`, `required`, dll)
- [ ] Remote template fetching (GitHub/GitLab)
- [ ] Local caching dengan TTL strategy

### Fase 4 — Distribution & Ecosystem
**Goal:** Siap untuk open source dan komunitas.

- [ ] Plugin system (executable-based renderer)
- [ ] `goreleaser` setup (binary untuk Linux, macOS, Windows, ARM)
- [ ] Install script (`curl -sSL https://... | sh`)
- [ ] Dokumentasi publik (website atau GitHub Pages)
- [ ] Template registry (opsional — GitHub Topics: `symphony-template`)
- [ ] GitHub Actions workflow untuk CI/CD Symphony sendiri

---

## 17. Architect's Notes & Warnings

### ⚠ Template Rot
Template yang tidak dirawat akan ketinggalan zaman seiring library melakukan breaking changes. **Solusi:** Command `symphony check` harus memeriksa versi minimum dependency yang diinjeksi ke `go.mod` atau `package.json` terhadap versi terbaru di registry publik.

### ⚠ Scope Creep di AST
Tentukan batas AST Injection lebih awal. Untuk mayoritas kasus pakai, injeksi berbasis anchor comment (regex-based) sudah cukup dan jauh lebih mudah diimplementasikan. Pendekatan `go/ast` penuh hanya diperlukan jika ada kebutuhan analisis semantik — misalnya "inject setelah deklarasi fungsi bernama X". Jangan membangun `go/ast` integration di Fase 1.

### ⚠ Expression Evaluator Security
Jangan pernah menggunakan `os/exec` atau evaluasi runtime (reflection) untuk mengevaluasi ekspresi kondisional `if:`. Implementasikan evaluator minimal sendiri atau gunakan library ringan seperti `gval`. Ini penting untuk menghindari template injection attacks ketika template diunduh dari sumber eksternal.

### ⚠ Concurrent File Writing
Saat menulis banyak file secara paralel (untuk performa), pastikan writer menggunakan goroutine dengan bounded semaphore — jangan unbounded goroutine spawning yang bisa menghabiskan file descriptor OS pada proyek besar.

### 💡 Rekomendasi untuk Open Source
Sebelum mempublikasikan ke open source, implementasikan setidaknya: schema validation, dry-run mode, dan lock file. Tiga fitur ini adalah yang paling banyak ditanyakan komunitas di tool-tool serupa dan akan menentukan kesan pertama kontributor terhadap kematangan proyek.

---

*Blueprint ini adalah living document. Update seiring perkembangan implementasi.*
