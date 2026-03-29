# Symphony CLI — Phase 00: Project Bootstrap & Skeleton
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 0 dari 5  
> **Tujuan:** Inisialisasi repositori, struktur folder kosong, konfigurasi tooling, dan dependency awal  
> **Prasyarat:** Go 1.22+ dan Git sudah terinstal di mesin  
> **Output yang diharapkan:** Repositori Go yang bisa di-build (`go build ./...`) meskipun belum melakukan apapun

---

## Konteks untuk AI

Kamu sedang membangun **Symphony CLI** — sebuah adaptive scaffolding engine berbasis Go. Tugas kamu di fase ini adalah **hanya** menyiapkan kerangka proyek: struktur folder, file kosong dengan package declaration yang benar, konfigurasi module, dan tooling. Tidak ada logika bisnis yang ditulis di fase ini. Setiap file cukup berisi deklarasi `package` dan komentar placeholder yang menjelaskan apa yang akan diisi di fase berikutnya.

Filosofi utama yang harus tercermin dalam struktur ini:
- **Data-Driven Core**: engine dan template sepenuhnya terpisah
- **Single binary distribution**: tidak ada dependency runtime eksternal
- **Separation of concerns**: setiap package memiliki tanggung jawab tunggal yang jelas

---

## Tugas 1 — Inisialisasi Go Module

Buat file `go.mod` dengan konfigurasi berikut:

```
module github.com/username/symphony

go 1.22
```

Kemudian jalankan perintah berikut untuk menambahkan seluruh dependency yang dibutuhkan selama pengembangan. Jelaskan setiap dependency yang ditambahkan dengan komentar di `go.mod`.

### Direct Dependencies yang Harus Ditambahkan

```
# CLI Framework
github.com/spf13/cobra v1.8.1
github.com/spf13/viper v1.18.2

# Terminal UI
github.com/charmbracelet/bubbletea v0.26.6
github.com/charmbracelet/bubbles v0.18.0
github.com/charmbracelet/lipgloss v0.10.0
github.com/charmbracelet/glamour v0.7.0

# YAML & Schema
gopkg.in/yaml.v3 v3.0.1
github.com/xeipuuv/gojsonschema v1.2.0

# Expression Evaluator (untuk kondisi `if:` di template)
github.com/nicholasgasior/gsfmt v0.0.0  # placeholder — ganti dengan gval
github.com/PaesslerAG/gval v1.2.2

# Networking
github.com/hashicorp/go-getter v1.7.4
github.com/google/go-github/v60 v60.0.0
github.com/mholt/archiver/v3 v3.5.1

# Testing
github.com/stretchr/testify v1.9.0
```

---

## Tugas 2 — Buat Seluruh Struktur Folder & File Kerangka

Buat **semua** folder dan file berikut. Setiap file `.go` harus berisi:
1. Deklarasi `package` yang sesuai
2. Komentar blok di bagian atas yang menjelaskan tanggung jawab file tersebut
3. Placeholder `// TODO: Implementasi di Phase XX` dengan nomor fase yang tepat

```
symphony/
│
├── main.go
│
├── cmd/
│   ├── root.go
│   ├── gen.go
│   ├── check.go
│   ├── list.go
│   ├── update.go
│   └── regen.go
│
├── internal/
│   ├── engine/
│   │   ├── engine.go
│   │   ├── context.go
│   │   ├── walker.go
│   │   ├── renderer.go
│   │   ├── writer.go
│   │   └── hooks.go
│   │
│   ├── blueprint/
│   │   ├── parser.go
│   │   ├── validator.go
│   │   ├── schema.go
│   │   └── resolver.go
│   │
│   ├── ast/
│   │   ├── injector_interface.go
│   │   ├── go_injector.go
│   │   └── ts_injector.go
│   │
│   ├── remote/
│   │   ├── fetcher.go
│   │   ├── github.go
│   │   └── cache.go
│   │
│   ├── tui/
│   │   ├── prompt.go
│   │   ├── select.go
│   │   ├── multiselect.go
│   │   ├── input.go
│   │   ├── confirm.go
│   │   ├── progress.go
│   │   └── summary.go
│   │
│   ├── lock/
│   │   ├── writer.go
│   │   └── reader.go
│   │
│   └── config/
│       └── config.go
│
├── pkg/
│   └── expr/
│       └── eval.go
│
├── testdata/
│   └── templates/
│       ├── simple-go/
│       │   └── template.yaml        ← file YAML kosong dengan komentar placeholder
│       └── hexagonal-go/
│           └── template.yaml        ← file YAML kosong dengan komentar placeholder
│
└── docs/
    ├── blueprint-spec.md
    └── contributing.md
```

---

## Tugas 3 — Isi `main.go`

`main.go` harus berisi hanya satu hal: memanggil fungsi `Execute()` dari package `cmd`. Ini adalah pola standar Cobra.

```go
// main.go
// Symphony CLI — The Adaptive Scaffolding Engine
// Entry point utama. Semua logika dimulai dari cmd/root.go
package main

import "github.com/username/symphony/cmd"

func main() {
    cmd.Execute()
}
```

---

## Tugas 4 — Setup `cmd/root.go` Sebagai Kerangka Cobra

Buat root command yang fungsional namun kosong. Root command adalah titik masuk semua sub-command. Di fase ini, cukup daftarkan nama aplikasi, versi, dan flag global. Sub-command belum perlu melakukan apapun.

File ini harus:
- Mendefinisikan `rootCmd` sebagai `*cobra.Command` dengan `Use`, `Short`, dan `Long` description
- Mendefinisikan fungsi `Execute()` yang dipanggil dari `main.go`
- Mendaftarkan global flags: `--verbose` (`-v`), `--config` (`-c`), `--format`
- Membaca konfigurasi dari `~/.symphony/config.yaml` menggunakan Viper
- Me-register semua sub-command dari file lain di folder `cmd/` (cukup dengan `rootCmd.AddCommand(...)`)

Contoh deklarasi global flags:
```go
rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Tampilkan log detail")
rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path ke config file kustom")
rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "human", "Format output: human | json")
```

---

## Tugas 5 — Buat Stub untuk Setiap Sub-command

Buat file untuk setiap command berikut. Setiap file harus mendefinisikan variabel `*cobra.Command` yang valid dan me-register dirinya ke `rootCmd` via fungsi `init()`. Command boleh menampilkan pesan `"[command] belum diimplementasikan"` saat dijalankan.

### `cmd/gen.go`
```go
// Perintah: symphony gen <source> [flags]
// Fungsi: Menjalankan scaffolding dari sebuah template source
// Flags: --out (-o), --dry-run (-d), --no-hooks, --yes (-y)
```

### `cmd/check.go`
```go
// Perintah: symphony check <source>
// Fungsi: Memvalidasi template tanpa menjalankan scaffolding
```

### `cmd/list.go`
```go
// Perintah: symphony list
// Fungsi: Menampilkan template lokal dan yang ter-cache
```

### `cmd/update.go`
```go
// Perintah: symphony update <source>
// Fungsi: Memperbarui template ke versi terbaru
```

### `cmd/regen.go`
```go
// Perintah: symphony re-gen [flags]
// Fungsi: Mengulang scaffold dari symphony.lock yang ada
// Prasyarat: harus dijalankan dari dalam direktori yang memiliki symphony.lock
```

---

## Tugas 6 — Buat `Makefile`

Buat `Makefile` di root proyek dengan target-target berikut. Semua target harus berfungsi meskipun sebagian besar kode belum diimplementasikan.

```makefile
.PHONY: build test lint clean run

# Nama binary output
BINARY_NAME=symphony
BUILD_DIR=./bin

build:
	@echo "Building Symphony CLI..."
	@go build -ldflags="-X 'github.com/username/symphony/cmd.Version=dev'" -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

run:
	@go run ./main.go $(ARGS)

test:
	@go test ./... -v -race

test-cover:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@golangci-lint run ./...

clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

# Install binary ke GOPATH/bin
install:
	@go install ./main.go

# Jalankan go mod tidy
tidy:
	@go mod tidy
```

---

## Tugas 7 — Buat `.goreleaser.yaml` Kerangka

Buat konfigurasi goreleaser yang akan digunakan di Fase 4 untuk distribusi binary. Di fase ini cukup buat kerangkanya saja.

```yaml
# .goreleaser.yaml
# Konfigurasi untuk build & release multi-platform
# Digunakan di Fase 4 — Distribution

version: 2

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/username/symphony/cmd.Version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

---

## Tugas 8 — Verifikasi Akhir Fase 0

Setelah semua file dibuat, jalankan perintah berikut dan pastikan **tidak ada error**:

```bash
# 1. Pastikan semua dependency terunduh
go mod tidy

# 2. Pastikan proyek bisa di-compile tanpa error
go build ./...

# 3. Pastikan binary bisa dijalankan dan menampilkan help
go run ./main.go --help

# 4. Pastikan semua sub-command terdaftar
go run ./main.go gen --help
go run ./main.go check --help
go run ./main.go list --help
```

Output yang diharapkan dari `--help` harus menampilkan nama aplikasi, deskripsi singkat, dan daftar sub-command yang tersedia — meskipun sub-command tersebut belum melakukan apapun.

---

## Checklist Selesai Fase 0

Sebelum melanjutkan ke Fase 1, pastikan semua kondisi berikut terpenuhi:

- [ ] `go build ./...` berjalan tanpa error atau warning
- [ ] `go run ./main.go --help` menampilkan output yang benar
- [ ] Semua 5 sub-command terdaftar dan menampilkan stub message saat dijalankan
- [ ] `go mod tidy` tidak menambah atau menghapus dependency
- [ ] Semua folder dalam struktur yang didefinisikan di Tugas 2 sudah ada
- [ ] Setiap file `.go` memiliki deklarasi `package` yang benar
- [ ] `Makefile` target `build` menghasilkan binary di `./bin/symphony`

---

## Catatan Penting untuk AI

- **Jangan** menulis logika bisnis apapun di fase ini. Jika ada godaan untuk langsung mengimplementasikan sesuatu, tahan dan buat komentar `TODO` saja.
- Semua import yang belum digunakan harus dihapus atau di-blank import (`_ "package"`) agar tidak menyebabkan compile error.
- Gunakan `//nolint:unused` hanya jika benar-benar diperlukan untuk menghindari lint error pada placeholder functions.
- Nama package harus mengikuti konvensi Go: lowercase, satu kata, sesuai nama folder.
