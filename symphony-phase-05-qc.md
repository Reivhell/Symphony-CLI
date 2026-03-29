# Symphony CLI — Phase 05: Quality Control & Release Readiness
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 5 dari 5 (Final)  
> **Tujuan:** Audit menyeluruh terhadap keamanan, performa, keandalan, dan kelayakan pakai sebelum Symphony dianggap siap untuk rilis publik  
> **Prasyarat:** Phase 04 selesai. Semua fitur terimplementasi dan bisa di-build untuk semua platform  
> **Output yang diharapkan:** Laporan QC lengkap, semua critical dan major issue terselesaikan, test coverage ≥ 75%, dan Symphony dinyatakan layak rilis

---

## Konteks untuk AI

Fase ini bukan fase pembangunan fitur. Ini adalah fase audit dan perbaikan. Tugasmu adalah mengevaluasi seluruh codebase Symphony secara sistematis dari empat dimensi: keamanan, performa, keandalan, dan pengalaman pengguna. Temukan masalah, laporkan dengan jelas, dan perbaiki secara terstruktur.

Setiap temuan harus dikategorikan sebagai **CRITICAL** (harus diperbaiki sebelum rilis), **MAJOR** (sangat disarankan diperbaiki), atau **MINOR** (improvement opsional). Tidak ada Symphony yang boleh dirilis jika masih ada issue berstatus CRITICAL yang terbuka.

---

## Tugas 1 — Security Audit

### 1.1 Expression Evaluator (`pkg/expr/eval.go`)

Verifikasi bahwa evaluator tidak bisa dieksploitasi melalui ekspresi yang dibuat secara jahat dalam template yang diunduh dari sumber tidak tepercaya.

Buat test cases berikut dan verifikasi bahwa semua mengembalikan error, bukan panic atau eksekusi yang tidak diharapkan:

```go
// Test cases yang HARUS mengembalikan error:
testCases := []string{
    `os.Exit(1)`,
    `exec.Command("rm", "-rf", "/")`,
    `__import__('os').system('whoami')`,
    strings.Repeat("A", 10000),  // Ekspresi sangat panjang — potensi ReDoS
    `1/0`,                        // Division by zero — harus error, bukan panic
    `null.field`,                  // Null dereference — harus error, bukan panic
}
```

### 1.2 Path Traversal Prevention (`internal/engine/writer.go`, `internal/remote/`)

Verifikasi bahwa template yang mengandung path berbahaya tidak bisa menulis file di luar `OutputDir`.

```go
// Test cases path traversal yang HARUS diblokir:
dangerousPaths := []string{
    "../../../etc/passwd",
    "../../.ssh/authorized_keys",
    "/absolute/path/outside/output",
    "normal/../../../evil",
    "..\\..\\Windows\\System32",  // Windows
}

// Untuk setiap path di atas, pastikan WriteFile mengembalikan error
// dan tidak membuat file di lokasi tersebut.
```

Implementasikan fungsi `sanitizePath` yang memverifikasi bahwa path yang dihasilkan setelah `filepath.Clean()` dan `filepath.Abs()` benar-benar berada di dalam `OutputDir`. Jika tidak, kembalikan error `ErrPathTraversal` tanpa menulis apapun.

### 1.3 Plugin Execution Safety (`internal/engine/plugin.go`)

Verifikasi bahwa plugin tidak bisa:
- Mengakses direktori di luar scope yang diizinkan
- Berjalan lebih lama dari timeout yang ditetapkan
- Menyebabkan panic pada Symphony jika plugin crash

Pastikan timeout implementation menggunakan `context.WithTimeout` dan bahwa cancellation benar-benar membunuh child process, bukan hanya menghentikan menunggu output-nya.

### 1.4 Template Injection di `RenderString`

Verifikasi bahwa nilai yang dimasukkan user melalui prompt tidak bisa mengandung Go template syntax yang dieksekusi pada layer berikutnya.

```go
// Skenario: user memasukkan nilai prompt yang mengandung template syntax
dangerousInput := "{{.SECRET_VAR}}"
ctx := &EngineContext{
    Values: map[string]any{
        "PROJECT_NAME": dangerousInput,
        "SECRET_VAR":   "sensitive-data",
    },
}
// Verifikasi bahwa rendering tidak mengekspos SECRET_VAR
```

Solusi: semua nilai yang berasal dari input user harus di-escape sebelum dimasukkan ke dalam context yang digunakan untuk rendering template. Nilai user tidak boleh diinterpretasikan sebagai template syntax.

---

## Tugas 2 — Performance Audit

### 2.1 Startup Time

Ukur waktu startup Symphony CLI tanpa melakukan operasi apapun:

```bash
# Ukur 10 kali dan ambil rata-rata
for i in $(seq 1 10); do
    time ./bin/symphony --help > /dev/null 2>&1
done
```

**Target:** Startup time harus di bawah 100ms di semua platform. Jika melebihi, investigasi penyebabnya. Kemungkinan penyebab umum: Viper yang memuat seluruh config file di init time, atau import dari package berat yang tidak diperlukan.

### 2.2 Benchmark Renderer

Buat benchmark test untuk mengukur performa rendering:

```go
// internal/engine/renderer_bench_test.go

func BenchmarkRenderFile_Simple(b *testing.B) {
    // Template sederhana dengan 5 variable substitutions
}

func BenchmarkRenderFile_Complex(b *testing.B) {
    // Template kompleks dengan conditionals, loops, dan 20+ variable substitutions
}

func BenchmarkRenderFile_Large(b *testing.B) {
    // File template besar (10KB+)
}
```

Jalankan dengan `go test -bench=. -benchmem ./internal/engine/...` dan dokumentasikan hasilnya. Identifikasi bottleneck jika ada.

### 2.3 File Writing Concurrency

Verifikasi bahwa bounded semaphore pada writer benar-benar membatasi concurrent goroutines. Buat test yang mencoba menulis 100 file secara bersamaan dan verifikasi bahwa maksimal 10 goroutine aktif pada satu waktu.

```go
func TestWriter_BoundedConcurrency(t *testing.T) {
    // Pantau peak goroutine count selama penulisan 100 file
    // Verifikasi tidak melebihi batas yang ditetapkan (10)
}
```

---

## Tugas 3 — Reliability & Error Handling Audit

### 3.1 Audit Semua Error Paths

Lakukan code review terhadap setiap error yang dikembalikan di seluruh codebase. Setiap error yang sampai ke user harus:

Pertama, memiliki pesan yang menjelaskan apa yang salah — bukan hanya "error" atau pesan dari library internal yang tidak bermakna bagi user.

Kedua, jika memungkinkan, menyertakan saran konkret tentang cara memperbaikinya.

Ketiga, tidak mengekspos informasi internal sistem seperti path absolut di mesin developer, stack trace, atau variabel internal.

Buat checklist dan verifikasi setiap error path di file-file berikut:
- `internal/blueprint/parser.go`
- `internal/remote/fetcher.go`
- `internal/engine/engine.go`
- `internal/ast/anchor_injector.go`

### 3.2 Interrupt Handling (Ctrl+C)

Test bahwa menekan Ctrl+C pada setiap titik dalam alur tidak meninggalkan state yang rusak:

Pertama, Ctrl+C saat prompting: tidak ada file yang dibuat, tidak ada direktori output kosong yang tertinggal, exit code 4.

Kedua, Ctrl+C saat generation sedang berjalan (di tengah proses): semua file yang sudah terlanjur ditulis harus dihapus (rollback), atau setidaknya ada pesan yang memberitahu user bahwa direktori output mungkin dalam keadaan partial dan perlu dibersihkan. Exit code 4.

Ketiga, Ctrl+C saat post-scaffold hook berjalan: proses hook diterminasi, pesan peringatan ditampilkan bahwa proyek mungkin membutuhkan `go mod tidy` atau perintah serupa secara manual. Exit code 4.

Implementasikan rollback mechanism untuk skenario kedua jika belum ada:

```go
// internal/engine/writer.go
// WrittenFiles melacak semua file yang sudah ditulis dalam sesi ini
// untuk keperluan rollback jika terjadi interrupt
type WriteSession struct {
    WrittenFiles []string
    mutex        sync.Mutex
}

func (s *WriteSession) Rollback() error {
    // Hapus semua file yang sudah ditulis, dalam urutan terbalik
}
```

### 3.3 Network Failure Handling

Test behavior Symphony ketika network tidak tersedia saat fetching remote template:

Scenario 1: DNS resolution failure — pesan error harus menyarankan untuk mengecek koneksi internet.

Scenario 2: Partial download (koneksi terputus di tengah) — Symphony harus menghapus partial download dan tidak menggunakan cache yang corrupt.

Scenario 3: Rate limiting dari GitHub API — Symphony harus menyarankan penggunaan `--github-token` atau menunggu beberapa saat.

Semua scenario ini harus menghasilkan pesan error yang actionable, bukan raw Go networking errors.

### 3.4 Concurrent Access ke Cache

Test bahwa dua instance Symphony yang berjalan bersamaan (misalnya dalam environment CI dengan parallel jobs) tidak merusak cache satu sama lain. Implementasikan file-based locking pada operasi cache write menggunakan `flock` atau pendekatan serupa.

---

## Tugas 4 — Test Coverage Audit

### 4.1 Generate Coverage Report

```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | tail -1  # Tampilkan total coverage
```

### 4.2 Coverage Targets per Package

Verifikasi bahwa setiap package memenuhi target coverage minimum berikut. Package yang tidak memenuhi target harus ditambahkan test-nya sebelum rilis.

| Package | Target Coverage | Alasan |
|---|---|---|
| `pkg/expr` | ≥ 90% | Komponen keamanan kritis |
| `internal/blueprint` | ≥ 85% | Parsing dan validasi adalah fondasi |
| `internal/engine` | ≥ 80% | Logika bisnis utama |
| `internal/ast` | ≥ 80% | Modifikasi file — berisiko tinggi jika salah |
| `internal/lock` | ≥ 85% | Reproducibility bergantung pada ini |
| `internal/remote` | ≥ 70% | Network code sulit ditest sepenuhnya |
| `internal/tui` | ≥ 50% | UI code lebih sulit ditest secara otomatis |

### 4.3 Test Case yang Wajib Ada

Jika test berikut belum ada, buat sebelum rilis:

```go
// Wajib ada di pkg/expr/eval_test.go
TestEvaluate_EmptyExpression_ReturnsTrue
TestEvaluate_CircularReference_ReturnsError  
TestEvaluate_VeryLongExpression_ReturnsError
TestEvaluate_InjectionAttempt_ReturnsError

// Wajib ada di internal/engine/writer_test.go  
TestWriteFile_PathTraversal_ReturnsError
TestWriteFile_DryRun_DoesNotWriteToDisk
TestWriteFile_ParentDirCreated_WhenNotExists

// Wajib ada di internal/blueprint/parser_test.go
TestParse_CircularInheritance_ReturnsError
TestParse_MissingRequiredField_ReturnsDescriptiveError
TestParse_IncompatibleSymphonyVersion_ReturnsError

// Wajib ada di internal/engine/engine_test.go (integration)
TestEngine_FullScaffold_WithAllFeatures
TestEngine_ConditionalActions_SkipsCorrectly
TestEngine_HookFailure_ReturnsError
TestEngine_RollbackOnInterrupt
```

---

## Tugas 5 — End-to-End Scenario Testing

Jalankan skenario berikut secara manual dan verifikasi hasilnya sesuai ekspektasi. Dokumentasikan hasil setiap skenario dalam format tabel di bawah.

### Skenario 1: Happy Path — Local Template

```bash
symphony gen ./testdata/templates/hexagonal-go \
  --out /tmp/e2e-test-1 \
  --yes
```

Verifikasi: semua file terbuat, `symphony.lock` ada, hooks dijalankan, exit code 0.

### Skenario 2: Dry-Run Mode

```bash
symphony gen ./testdata/templates/hexagonal-go \
  --out /tmp/e2e-test-2 \
  --dry-run
```

Verifikasi: tidak ada file yang dibuat di `/tmp/e2e-test-2`, exit code 0.

### Skenario 3: Conditional Logic

Gunakan template dengan conditional action. Jawab prompt sehingga salah satu kondisi tidak terpenuhi. Verifikasi bahwa file yang seharusnya di-skip benar-benar tidak ada di output.

### Skenario 4: Re-generation

```bash
# Setup: jalankan gen terlebih dahulu
symphony gen ./testdata/templates/hexagonal-go --out /tmp/e2e-test-4 --yes

# Hapus beberapa file
rm /tmp/e2e-test-4/cmd/main.go

# Re-generate dari lock file
cd /tmp/e2e-test-4 && symphony re-gen --yes
```

Verifikasi: file yang dihapus kembali ada, kontennya identik dengan yang pertama.

### Skenario 5: Invalid Input Validation

Jalankan `symphony gen` dan masukkan nilai yang tidak valid untuk field dengan validasi `regex`. Verifikasi: error ditampilkan inline dengan pesan yang jelas dan saran perbaikan, program tidak crash.

### Skenario 6: Template dengan Plugin

Buat template sederhana yang mendaftarkan plugin custom dan verifikasi bahwa plugin dipanggil dengan benar dan output-nya digunakan.

### Skenario 7: `symphony check` pada Template Valid dan Invalid

```bash
symphony check ./testdata/templates/hexagonal-go  # Harus lulus
symphony check ./testdata/templates/broken-yaml    # Harus gagal dengan pesan deskriptif
```

### Template Laporan Skenario

Isi tabel ini setelah menjalankan setiap skenario:

| Skenario | Status | Exit Code | Catatan |
|---|---|---|---|
| 1: Happy Path | ✔/✖ | | |
| 2: Dry-Run | ✔/✖ | | |
| 3: Conditional | ✔/✖ | | |
| 4: Re-generation | ✔/✖ | | |
| 5: Invalid Input | ✔/✖ | | |
| 6: Plugin | ✔/✖ | | |
| 7a: Check Valid | ✔/✖ | | |
| 7b: Check Invalid | ✔/✖ | | |

---

## Tugas 6 — Platform Compatibility Verification

Verifikasi bahwa binary yang dihasilkan goreleaser berfungsi dengan benar di semua platform target. Gunakan virtual machine atau container jika platform fisik tidak tersedia.

### Linux (amd64)
```bash
# Gunakan Docker jika tidak punya Linux
docker run --rm -v $(pwd)/dist:/app ubuntu:22.04 /app/symphony_linux_amd64/symphony version
docker run --rm -v $(pwd)/dist:/app ubuntu:22.04 /app/symphony_linux_amd64/symphony gen --help
```

### macOS (Intel + Apple Silicon)
Jalankan langsung di hardware macOS. Verifikasi bahwa binary arm64 berjalan di Apple Silicon tanpa Rosetta 2.

### Windows (amd64)
Jalankan di Windows native atau VM. Perhatikan khusus path separator (`\` vs `/`) — Symphony harus menormalkan semua path ke slash sebelum diproses.

---

## Tugas 7 — Documentation Final Review

### 7.1 README Accuracy Check

Baca seluruh README dan verifikasi bahwa:
- Semua command yang disebutkan benar-benar ada dan berfungsi
- Semua flags yang disebutkan sudah diimplementasikan
- Quick start guide bisa diikuti dari awal sampai akhir oleh seseorang yang belum pernah melihat proyek ini
- Install script URL yang disebutkan benar

### 7.2 Blueprint Spec Completeness Check

Verifikasi bahwa `docs/blueprint-spec.md` mendokumentasikan **semua** field yang ada dalam struct `Blueprint` di `internal/blueprint/schema.go`. Tidak boleh ada field yang terimplementasi tetapi tidak terdokumentasi.

### 7.3 Error Messages Review

Kumpulkan semua pesan error yang bisa ditampilkan ke user (semua string yang diteruskan ke `PrintError` atau diemit sebagai JSON error event). Verifikasi bahwa:
- Tidak ada pesan dalam bahasa Inggris yang campur dengan bahasa lain secara inkonsisten
- Tidak ada pesan yang mengandung informasi internal debugging yang tidak berguna bagi user
- Setiap pesan error memiliki saran yang actionable

---

## Tugas 8 — Release Readiness Checklist Final

Ini adalah gerbang terakhir sebelum rilis. Semua item harus dicentang.

### Kode & Kualitas
- [ ] `go build ./...` berjalan tanpa warning di semua platform
- [ ] `go test ./... -race` lulus 100% tanpa flaky tests
- [ ] `golangci-lint run ./...` tidak ada error (warning boleh ada jika didokumentasikan)
- [ ] Total test coverage ≥ 75%
- [ ] Semua package mencapai target coverage minimum masing-masing (Tugas 4.2)

### Keamanan
- [ ] Path traversal test lulus
- [ ] Expression injection test lulus
- [ ] Plugin timeout berfungsi
- [ ] Template injection test lulus
- [ ] Tidak ada hardcoded credentials atau secrets di codebase

### Keandalan
- [ ] Ctrl+C pada setiap fase menghasilkan clean exit dengan exit code 4
- [ ] Network failure menghasilkan pesan error yang actionable
- [ ] Partial write saat interrupt tidak meninggalkan state corrupt permanen
- [ ] Semua 8 skenario E2E lulus

### Platform
- [ ] Binary Linux amd64 berfungsi
- [ ] Binary Linux arm64 berfungsi
- [ ] Binary macOS amd64 berfungsi
- [ ] Binary macOS arm64 berfungsi
- [ ] Binary Windows amd64 berfungsi
- [ ] Checksum file tersedia dan benar

### Distribusi
- [ ] Install script berfungsi di Linux
- [ ] Install script berfungsi di macOS
- [ ] `symphony version` menampilkan versi yang benar di semua platform
- [ ] GitHub Actions CI pipeline hijau di semua matrix

### Dokumentasi
- [ ] README quick start dapat diikuti dari awal sampai akhir
- [ ] Semua field `template.yaml` terdokumentasi di blueprint spec
- [ ] CHANGELOG untuk v0.1.0 sudah ditulis
- [ ] LICENSE file ada (MIT atau Apache 2.0)

---

## Format Laporan QC Final

Setelah menyelesaikan semua tugas di atas, hasilkan laporan QC dalam format berikut:

```markdown
# Symphony CLI — QC Report
**Tanggal:** YYYY-MM-DD  
**Versi:** v0.1.0  
**Reviewer:** [AI/Nama]

## Ringkasan

| Kategori | Status | Issue Ditemukan | Issue Diselesaikan |
|---|---|---|---|
| Security | PASS/FAIL | N | N |
| Performance | PASS/FAIL | N | N |
| Reliability | PASS/FAIL | N | N |
| Test Coverage | PASS/FAIL | N | N |
| E2E Scenarios | PASS/FAIL | N/8 | N/8 |
| Platform Compat | PASS/FAIL | N | N |
| Documentation | PASS/FAIL | N | N |

## Issue yang Ditemukan

### CRITICAL (harus diselesaikan)
- [Deskripsi issue, file yang terdampak, cara memperbaiki]

### MAJOR (sangat disarankan)
- [...]

### MINOR (opsional)
- [...]

## Verdict

**LAYAK RILIS / TIDAK LAYAK RILIS**

Alasan: [penjelasan singkat]
```

---

## Catatan Penting untuk AI

QC adalah tentang menemukan masalah yang tidak terlihat saat pembangunan, bukan mengkonfirmasi bahwa semua baik-baik saja. Approach dengan mindset adversarial: anggap setiap komponen mungkin memiliki bug dan cari bukti sebaliknya, bukan sebaliknya. Temuan yang paling berharga adalah yang muncul dari kombinasi fitur — misalnya, apa yang terjadi ketika template inheritance dikombinasikan dengan remote fetching dan kondisi jaringan yang buruk secara bersamaan? Skenario kombinasi seperti ini sering terlewat dalam unit test biasa.
