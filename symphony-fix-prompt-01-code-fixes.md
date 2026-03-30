# Symphony CLI — Fix Prompt 01: Critical Code Fixes
> **Tipe Dokumen:** AI Execution Prompt
> **Sesi:** 1 dari 2
> **Cakupan:** Semua perbaikan pada level kode — build failure, module path, AST bug, Go version, expr evaluator, lock file hardening, Makefile
> **Referensi:** `symphony-remediation-blueprint.md` — Issue #1, #2, #3, #9, #11, Improvisasi 16.1, 16.3
> **Commit Target:** `chore/fix-01-critical-code-fixes`
> **Prasyarat:** Repositori `Reivhell/Symphony-CLI` sudah di-clone secara lokal

---

## Konteks untuk AI

Kamu adalah senior Go engineer yang bertugas memperbaiki repositori **Symphony CLI** (`github.com/Reivhell/Symphony-CLI`) sebelum development feature dilanjutkan. Repositori ini saat ini tidak bisa di-build karena beberapa masalah kritis yang akan kamu tangani secara menyeluruh dalam sesi ini.

Aturan yang wajib dipatuhi selama sesi ini:

Pertama, kerjakan setiap tugas **secara berurutan** sesuai nomor. Jangan melompat karena setiap tugas bergantung pada tugas sebelumnya. Kedua, setelah setiap tugas selesai, jalankan verifikasi command yang tertera dan pastikan output sesuai sebelum melanjutkan. Ketiga, **jangan** menulis kode fitur baru di luar perbaikan yang diminta. Keempat, semua perubahan dikumpulkan dan di-commit di akhir sesi dalam satu commit yang bersih.

---

## Tugas 1 — Perbaiki Build Failure di `internal/remote`

**Masalah:** Package `internal/remote` tidak bisa dikompilasi karena unused import `"os"` di `cache_test.go`, menyebabkan `go build ./...` dan `go test ./...` gagal secara menyeluruh.

**Bukti:**
```
internal\remote\cache_test.go:4:2: "os" imported and not used
FAIL  github.com/username/symphony/internal/remote [build failed]
```

**Yang harus dilakukan:**

Buka file `internal/remote/cache_test.go`. Temukan import `"os"` di blok import dan hapus baris tersebut. Jika file tersebut memiliki test functions yang kosong atau stub, pastikan setiap function tetap bisa dikompilasi. Jika ada test function yang membutuhkan `os` tetapi belum diimplementasikan, ganti dengan placeholder yang di-skip:

```go
package remote_test

import (
	"testing"
	// Hapus: "os" — tidak digunakan
)

// Untuk setiap test yang belum diimplementasikan, gunakan format ini:
func TestCache_Placeholder(t *testing.T) {
	t.Skip("TODO: implementasi di Phase 03")
}
```

**Verifikasi setelah selesai:**
```bash
go build ./internal/remote/...
# Expected: tidak ada output (silent success)

go test ./internal/remote/... -v
# Expected: semua test PASS atau SKIP, tidak ada FAIL
```

---

## Tugas 2 — Rename Module Path dari Placeholder ke Identitas Nyata

**Masalah:** `go.mod`, `main.go`, `.goreleaser.yaml`, dan kemungkinan seluruh file `.go` masih menggunakan `github.com/username/symphony` sebagai module path. Ini menyebabkan inkonsistensi dengan lokasi repositori yang sebenarnya di GitHub.

**Module path target:** `github.com/Reivhell/symphony`

**Yang harus dilakukan:**

Langkah 2a — Jalankan rename menyeluruh menggunakan `sed` pada semua file yang relevan:

```bash
# Temukan semua file yang mengandung placeholder
grep -rl "github.com/username/symphony" . \
  --include="*.go" \
  --include="*.yaml" \
  --include="*.yml" \
  --include="*.md" \
  --include="*.sh" \
  --exclude-dir=".git"
```

Langkah 2b — Lakukan substitusi di semua file tersebut:

```bash
# Linux/macOS:
find . -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.sh" \) \
  -not -path "./.git/*" \
  -exec sed -i 's|github.com/username/symphony|github.com/Reivhell/symphony|g' {} +

# Update go.mod secara eksplisit untuk keamanan:
go mod edit -module github.com/Reivhell/symphony
```

Langkah 2c — Jalankan `go mod tidy` untuk menyinkronkan semua dependency:

```bash
go mod tidy
```

**Verifikasi setelah selesai:**
```bash
# Harus tidak menghasilkan output apapun (zero hits)
grep -r "username/symphony" . \
  --include="*.go" \
  --include="*.yaml" \
  --include="*.sh" \
  --exclude-dir=".git"

# Build harus tetap bersih
go build ./...
```

---

## Tugas 3 — Turunkan Go Version di `go.mod`

**Masalah:** `go.mod` mendeklarasikan `go 1.24.2` yang terlalu baru untuk GitHub Actions CI runners standar dan dapat menyebabkan masalah pada mesin developer yang menggunakan versi Go yang lebih lama.

**Yang harus dilakukan:**

Langkah 3a — Ubah versi Go di `go.mod`:

```bash
go mod edit -go 1.22
go mod tidy
```

Langkah 3b — Buat file `.go-version` di root repositori dengan konten berikut. File ini dibaca oleh `goenv`, `mise`, dan beberapa CI systems untuk secara otomatis menggunakan versi yang tepat:

```
1.22.10
```

**Verifikasi setelah selesai:**
```bash
# go.mod harus menampilkan "go 1.22"
head -5 go.mod

# Build harus tetap bersih — jika ada fitur Go 1.23+ yang tidak sengaja digunakan, akan muncul error di sini
go build ./...
go test ./...
```

---

## Tugas 4 — Perbaiki Bug Logika di AST Injector

**Masalah:** `TestGoInjector_SyntaxValidFormatter` gagal karena strategi `after-anchor` di `internal/ast/anchor_injector.go` menyisipkan konten di posisi yang **terbalik** — konten yang diinjeksi muncul sebelum baris berikutnya, bukan setelahnya.

**Bukti dari `debug.log`:**
```
Error: "...fmt.Println(\"Injected Route\")\n\tfmt.Println(\"Start\")..."
       does not contain
       "\tfmt.Println(\"Start\")\n\tfmt.Println(\"Injected Route\")"
```

Artinya: `Injected Route` muncul **sebelum** `Start` di output, padahal seharusnya **sesudah**.

**Yang harus dilakukan:**

Langkah 4a — Buka `internal/ast/anchor_injector.go` dan temukan implementasi strategi `after-anchor`. Bug hampir pasti ada pada bagian yang menyambungkan kembali string setelah anchor ditemukan. Pola yang salah biasanya terlihat seperti ini:

```go
// POLA YANG SALAH — menggunakan SplitN yang menempatkan
// injectedContent sebelum sisa konten file
parts := strings.SplitN(content, anchor, 2)
result := parts[0] + anchor + "\n" + injectedContent + parts[1]
// Masalah: parts[1] sudah mengandung "\nfmt.Println(\"Start\")"
// sehingga injectedContent disisipkan SEBELUM "Start"
```

Langkah 4b — Ganti dengan implementasi yang benar menggunakan pendekatan line-by-line:

```go
// IMPLEMENTASI YANG BENAR untuk strategi "after-anchor"
func injectAfterAnchor(content string, anchor string, injectedContent string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)+1)

	for _, line := range lines {
		result = append(result, line)
		// Sisipkan konten tepat setelah baris yang mengandung anchor
		if strings.Contains(line, anchor) {
			result = append(result, injectedContent)
		}
	}
	return strings.Join(result, "\n")
}
```

Langkah 4c — Verifikasi bahwa implementasi yang sama juga benar untuk strategi `before-anchor` dan `replace-anchor`. Untuk `before-anchor`, konten diinjeksi **sebelum** baris anchor disisipkan ke `result`. Untuk `replace-anchor`, baris anchor tidak disisipkan ke `result`, hanya `injectedContent` yang ditambahkan di posisi tersebut.

Langkah 4d — Tambahkan properti opsional `remove_anchor` sebagai improvisasi. Jika `remove_anchor: true` di action, baris anchor tidak dimasukkan ke output akhir setelah injeksi:

```go
// Di struct Action (internal/blueprint/schema.go), tambahkan field:
RemoveAnchor bool `yaml:"remove_anchor"`

// Di injector, terapkan logika:
if strings.Contains(line, action.Anchor) {
    if action.Strategy == "after-anchor" {
        if !action.RemoveAnchor {
            result = append(result, line) // Pertahankan baris anchor
        }
        result = append(result, action.Content) // Sisipkan konten injeksi
    }
}
```

**Verifikasi setelah selesai:**
```bash
go test ./internal/ast/... -v -race
# Expected: semua test PASS termasuk TestGoInjector_SyntaxValidFormatter
# Tidak boleh ada satu pun FAIL
```

---

## Tugas 5 — Audit dan Hardening Expression Evaluator

**Masalah:** File `expr_test.log` yang ter-commit mengindikasikan bahwa test expression evaluator pernah dijalankan dan hasilnya perlu di-debug. Package `pkg/expr` adalah komponen keamanan kritis yang memproses ekspresi `if:` dari template eksternal — bug di sini memiliki implikasi keamanan.

**Yang harus dilakukan:**

Langkah 5a — Jalankan test yang sudah ada dan perhatikan hasilnya:

```bash
go test ./pkg/expr/... -v -race -count=1
```

Jika ada test yang gagal, perbaiki sebelum melanjutkan ke langkah berikutnya.

Langkah 5b — Tambahkan test khusus untuk security boundaries di `pkg/expr/eval_test.go`. Buat test function baru berikut:

```go
func TestEvaluate_SecurityBoundaries(t *testing.T) {
	// Semua ekspresi berikut harus mengembalikan error, BUKAN panic
	dangerousCases := []struct {
		name string
		expr string
	}{
		{"os exit attempt", `os.Exit(1)`},
		{"exec command attempt", `exec.Command("whoami")`},
		{"very long expression", strings.Repeat("A", 100_000)},
		{"division by zero", `1/0`},
		{"null dereference", `nil.field`},
		{"empty string", ``},
	}

	ctx := map[string]any{"SAFE_VAR": "value"}

	for _, tc := range dangerousCases {
		t.Run(tc.name, func(t *testing.T) {
			// Gunakan recover untuk memastikan tidak ada panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Evaluate panicked for expression %q: %v", tc.expr, r)
				}
			}()

			if tc.expr == "" {
				// Ekspresi kosong harus mengembalikan true (no condition = always run)
				result, err := Evaluate(tc.expr, ctx)
				assert.NoError(t, err)
				assert.True(t, result)
			} else {
				_, err := Evaluate(tc.expr, ctx)
				// Harus mengembalikan error, tidak boleh panic
				assert.NotNil(t, err, "expression %q should return error", tc.expr)
			}
		})
	}
}

func TestEvaluate_TemplateInjectionPrevention(t *testing.T) {
	// Nilai dari input user tidak boleh bisa mereferensikan variable lain
	ctx := map[string]any{
		"PROJECT_NAME": "{{.SECRET_VAR}}", // User memasukkan template syntax
		"SECRET_VAR":   "sensitive-data",
	}

	// Evaluasi ekspresi yang legitimate harus tetap bekerja
	result, err := Evaluate("PROJECT_NAME != ''", ctx)
	assert.NoError(t, err)
	assert.True(t, result)

	// Nilai user tidak boleh di-evaluate sebagai template
	result2, err2 := Evaluate("PROJECT_NAME == 'sensitive-data'", ctx)
	assert.NoError(t, err2)
	assert.False(t, result2, "template injection should not expose SECRET_VAR value")
}
```

Langkah 5c — Pastikan coverage `pkg/expr` mencapai minimal 85%:

```bash
go test ./pkg/expr/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "pkg/expr"
# Target: coverage >= 85%
```

**Verifikasi setelah selesai:**
```bash
go test ./pkg/expr/... -v -race
# Expected: semua test PASS, termasuk security boundary tests
```

---

## Tugas 6 — Tambahkan `file_checksums` ke Lock File

**Motivasi:** Field `file_checksums` di `symphony.lock` dibutuhkan oleh fitur Diff & Merge Engine yang direncanakan untuk v0.6.0. Menambahkannya sekarang tidak membutuhkan perubahan besar, namun lock file yang dihasilkan oleh versi ini akan langsung kompatibel dengan Merge Engine tanpa migrasi.

**Yang harus dilakukan:**

Langkah 6a — Buka `internal/lock/writer.go`. Temukan struct `LockFile` dan tambahkan field `FileChecksums`:

```go
// LockFile merepresentasikan konten dari file symphony.lock
type LockFile struct {
	SymphonyVersion string            `json:"symphony_version"`
	GeneratedAt     time.Time         `json:"generated_at"`
	Template        TemplateLockInfo  `json:"template"`
	Inputs          map[string]any    `json:"inputs"`
	OutputChecksum  string            `json:"output_checksum"`

	// FileChecksums merekam SHA-256 dari setiap file yang di-generate.
	// Digunakan oleh Diff & Merge Engine (v0.6.0) untuk mendeteksi
	// modifikasi manual setelah scaffold.
	// Key: path relatif dari output dir. Value: "sha256:<hex>".
	FileChecksums map[string]string `json:"file_checksums,omitempty"`
}
```

Langkah 6b — Tambahkan helper function untuk menghitung checksum sebuah file:

```go
// fileChecksum menghitung SHA-256 dari konten file dan mengembalikannya
// dalam format "sha256:<hex>".
func fileChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", hash)
}
```

Langkah 6c — Update fungsi `Write` di `writer.go` untuk menerima dan menyertakan `FileChecksums` dalam lock file. Pastikan field ini diisi oleh engine setelah semua file selesai ditulis.

Langkah 6d — Update `internal/engine/engine.go` untuk mengumpulkan checksum setiap file yang ditulis dan meneruskannya ke `lock.Write()`.

**Verifikasi setelah selesai:**
```bash
go build ./internal/lock/...
go test ./internal/lock/... -v
# Pastikan lock file yang dihasilkan mengandung field file_checksums dalam format JSON yang benar
```

---

## Tugas 7 — Update `Makefile` dengan Target Standar

**Yang harus dilakukan:**

Ganti atau perbarui `Makefile` di root repositori dengan konfigurasi berikut. Perhatikan bahwa indentasi pada `Makefile` menggunakan **tab**, bukan spasi:

```makefile
# ──────────────────────────────────────────────────────────────────────────────
# Symphony CLI — Makefile
# Gunakan 'make help' untuk melihat semua target yang tersedia.
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: all check build test lint vet clean install tidy snapshot help

# Build variables — di-inject ke binary via ldflags
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE      := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE    := github.com/Reivhell/symphony
LDFLAGS   := -ldflags="-s -w \
               -X '$(MODULE)/cmd.Version=$(VERSION)' \
               -X '$(MODULE)/cmd.Commit=$(COMMIT)' \
               -X '$(MODULE)/cmd.BuildDate=$(DATE)'"
BINARY    := ./bin/symphony

## all: Jalankan check lengkap (default target)
all: check

## check: Jalankan vet + build + test — wajib dijalankan sebelum push
check: vet build test
	@echo ""
	@echo "  ✔ All checks passed. Ready to push."
	@echo ""

## build: Kompilasi binary ke ./bin/symphony
build:
	@echo "Building $(VERSION)..."
	@mkdir -p ./bin
	@go build $(LDFLAGS) -o $(BINARY) ./main.go
	@echo "  ✔ Built: $(BINARY)"

## test: Jalankan seluruh test suite dengan race detector
test:
	@echo "Running tests..."
	@go test ./... -race -count=1

## test-cover: Jalankan test dan hasilkan coverage report
test-cover:
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "  Coverage report: coverage.html"

## vet: Jalankan go vet
vet:
	@go vet ./...

## lint: Jalankan golangci-lint
lint:
	@golangci-lint run ./...

## clean: Hapus semua build artifacts dan log files
clean:
	@rm -rf ./bin ./dist coverage.out coverage.html
	@find . -name "*.log" -not -path "./.git/*" -delete
	@echo "  ✔ Clean complete."

## install: Install binary ke GOPATH/bin
install:
	@go install $(LDFLAGS) ./main.go

## tidy: Jalankan go mod tidy
tidy:
	@go mod tidy

## snapshot: Build snapshot lokal via goreleaser (tanpa publish)
snapshot:
	@goreleaser release --snapshot --clean

## check-placeholders: Verifikasi tidak ada placeholder yang tertinggal
check-placeholders:
	@if grep -rn "username/symphony" \
	     --include="*.go" --include="*.yaml" --include="*.sh" \
	     --exclude-dir=".git" . 2>/dev/null; then \
	  echo "  ✖ ERROR: Placeholder strings found!" && exit 1; \
	fi
	@echo "  ✔ No placeholder strings found."

## help: Tampilkan daftar semua target yang tersedia
help:
	@echo "Symphony CLI — Available Makefile targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
```

**Verifikasi setelah selesai:**
```bash
make vet
make build
make test
make check
# Expected: semua target berjalan tanpa error
```

---

## Tugas 8 — Commit Semua Perubahan Sesi Ini

Setelah semua tugas di atas selesai dan terverifikasi, kumpulkan semua perubahan dalam commit yang bersih:

```bash
# Pastikan tidak ada yang tertinggal
git status

# Stage semua perubahan
git add -A

# Verifikasi satu kali lagi sebelum commit
go build ./...
go test ./... -race

# Commit dengan pesan Conventional Commits yang deskriptif
git commit -m "fix: resolve critical code issues before feature development

- fix: remove unused 'os' import in internal/remote/cache_test.go
- chore: rename module path from github.com/username/symphony to github.com/Reivhell/symphony
- chore: downgrade go version to 1.22 in go.mod, add .go-version file
- fix: correct after-anchor injection order in anchor_injector.go
- feat: add remove_anchor option to AST injection actions
- test: add security boundary tests for expression evaluator
- feat: add file_checksums field to LockFile struct for future Merge Engine
- chore: update Makefile with comprehensive targets and check-placeholders"
```

---

## Checklist Penyelesaian Sesi 1

Sebelum menutup sesi ini, verifikasi semua kondisi berikut terpenuhi:

- [ ] `go build ./...` lulus tanpa error atau warning
- [ ] `go test ./... -race` — semua test PASS atau SKIP, tidak ada FAIL
- [ ] `go vet ./...` — tidak ada output (clean)
- [ ] `grep -r "username/symphony" . --include="*.go"` — tidak ada hasil
- [ ] `TestGoInjector_SyntaxValidFormatter` — PASS
- [ ] `TestEvaluate_SecurityBoundaries` — semua subtest PASS
- [ ] `symphony.lock` yang dihasilkan mengandung field `file_checksums`
- [ ] `make check` — berjalan hijau
- [ ] Semua perubahan sudah di-commit dengan pesan yang tepat

Setelah semua item di atas dicentang, lanjutkan ke **Prompt 02: Repository, Infrastructure & Governance Fixes**.

---

## Catatan Penting untuk AI

Jangan mengubah logika fitur yang tidak berhubungan dengan perbaikan yang diminta. Jika selama proses investigasi kamu menemukan bug atau masalah lain di luar scope yang tertera, **catat sebagai komentar TODO** di kode yang bersangkutan dan dokumentasikan di akhir sesi — jangan perbaiki dalam sesi ini karena bisa mengubah scope commit dan menyulitkan review. Fokus adalah membuat build bersih dan test hijau, tidak lebih dari itu.



Use Context7  For Newest Knowledge