# Symphony CLI — Phase 01: Core Engine
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 1 dari 5  
> **Tujuan:** Implementasi seluruh logika inti — blueprint parser, expression evaluator, template renderer, file writer, hook runner, dan lock file generator  
> **Prasyarat:** Phase 00 selesai. `go build ./...` berjalan tanpa error  
> **Output yang diharapkan:** `symphony gen ./path/to/template --out ./output` berjalan fungsional via prompt terminal standar (stdin), tanpa TUI. Semua logika bisnis inti sudah bekerja.

---

## Konteks untuk AI

Fase ini adalah jantung dari Symphony CLI. Kamu mengimplementasikan semua logika yang membuat Symphony berbeda dari sekadar "copy-paster": parsing blueprint YAML, evaluasi kondisi `if:`, rendering template, dan penulisan file ke disk. TUI yang cantik belum dibangun di sini — fokus sepenuhnya pada **correctness** logika bisnis.

Urutan implementasi sangat penting. Ikuti urutan tugas di dokumen ini karena setiap tugas bergantung pada tugas sebelumnya.

---

## Tugas 1 — Definisikan Tipe Data Utama (`internal/blueprint/`)

Sebelum menulis parser, definisikan terlebih dahulu semua struct yang merepresentasikan struktur `template.yaml`. Ini adalah kontrak data yang akan digunakan oleh seluruh sistem.

### `internal/blueprint/schema.go`

Definisikan struct-struct berikut beserta tag YAML dan JSON-nya:

```go
// Blueprint adalah representasi lengkap dari satu file template.yaml
type Blueprint struct {
    SchemaVersion      string            `yaml:"schema_version" json:"schema_version"`
    Name               string            `yaml:"name" json:"name"`
    Version            string            `yaml:"version" json:"version"`
    Author             string            `yaml:"author" json:"author"`
    Description        string            `yaml:"description" json:"description"`
    MinSymphonyVersion string            `yaml:"min_symphony_version" json:"min_symphony_version"`
    Tags               []string          `yaml:"tags" json:"tags"`
    Extends            string            `yaml:"extends" json:"extends"`
    Validations        []ValidationRule  `yaml:"validations" json:"validations"`
    Prompts            []Prompt          `yaml:"prompts" json:"prompts"`
    Actions            []Action          `yaml:"actions" json:"actions"`
    CompletionMessage  string            `yaml:"completion_message" json:"completion_message"`
    Plugins            []Plugin          `yaml:"plugins" json:"plugins"`
}

// Prompt merepresentasikan satu pertanyaan interaktif
type Prompt struct {
    ID        string   `yaml:"id"`
    Question  string   `yaml:"question"`
    Type      string   `yaml:"type"`      // "input" | "select" | "confirm" | "multiselect"
    Options   []string `yaml:"options"`
    Default   any      `yaml:"default"`
    DependsOn string   `yaml:"depends_on"` // Ekspresi kondisional
}

// Action merepresentasikan satu instruksi yang akan dieksekusi engine
type Action struct {
    Type       string `yaml:"type"`       // "render" | "shell" | "ast-inject"
    Source     string `yaml:"source"`
    Target     string `yaml:"target"`
    If         string `yaml:"if"`         // Ekspresi kondisional opsional
    Command    string `yaml:"command"`    // Untuk type "shell"
    WorkingDir string `yaml:"working_dir"`
    Strategy   string `yaml:"strategy"`   // Untuk type "ast-inject"
    Anchor     string `yaml:"anchor"`
    Content    string `yaml:"content"`
}

// ValidationRule mendefinisikan aturan validasi untuk satu field prompt
type ValidationRule struct {
    Field   string `yaml:"field"`
    Rule    string `yaml:"rule"`    // "regex" | "required" | "min_length"
    Pattern string `yaml:"pattern"`
    Message string `yaml:"message"`
}

// Plugin mendefinisikan custom renderer eksternal
type Plugin struct {
    Name       string   `yaml:"name"`
    Executable string   `yaml:"executable"`
    Handles    []string `yaml:"handles"`
}
```

---

## Tugas 2 — Blueprint Parser (`internal/blueprint/parser.go`)

Implementasikan fungsi untuk membaca dan mem-parse file `template.yaml` dari path yang diberikan.

### Spesifikasi Fungsi

```go
// Parse membaca file template.yaml dari path yang diberikan,
// memvalidasi strukturnya, dan mengembalikan Blueprint yang sudah di-parse.
// Mengembalikan error jika file tidak ditemukan, tidak valid YAML,
// atau tidak sesuai dengan schema yang diharapkan.
func Parse(templateDir string) (*Blueprint, error)
```

### Logika yang Harus Diimplementasikan

Fungsi `Parse` harus melakukan langkah-langkah ini secara berurutan. Jika salah satu gagal, kembalikan error yang deskriptif.

Pertama, buka dan baca file `template.yaml` dari `templateDir`. Berikan error yang jelas jika file tidak ditemukan, bukan hanya "file not found" dari OS.

Kedua, unmarshal konten YAML ke dalam struct `Blueprint`. Tangkap semua YAML parsing error dan tambahkan konteks (misalnya nomor baris jika tersedia dari yaml.v3).

Ketiga, jalankan validasi dasar: pastikan field `name`, `version`, dan minimal satu `action` ada. Field kosong harus menghasilkan error yang menyebutkan nama field yang kosong.

Keempat, validasi `schema_version`. Jika field ini ada, pastikan nilainya adalah `"2"`. Jika bukan, kembalikan error yang menjelaskan format yang diharapkan.

Kelima, validasi `min_symphony_version` jika ada. Parse sebagai semver dan bandingkan dengan versi Symphony yang sedang berjalan. Jika template membutuhkan versi yang lebih tinggi, kembalikan error yang memberitahu user versi mana yang dibutuhkan.

---

## Tugas 3 — Expression Evaluator (`pkg/expr/eval.go`)

Ini adalah komponen keamanan kritis. Evaluator digunakan untuk mengevaluasi ekspresi kondisional pada field `if:` dan `depends_on:` di `template.yaml`.

### Aturan Keamanan yang Tidak Boleh Dilanggar

Jangan pernah menggunakan `os/exec`, `reflect`, atau `eval` untuk mengevaluasi ekspresi. Gunakan library `github.com/PaesslerAG/gval` sebagai evaluator yang aman.

### Spesifikasi Fungsi

```go
// Evaluate mengevaluasi ekspresi string terhadap context values yang diberikan.
// Mengembalikan boolean hasil evaluasi dan error jika ekspresi tidak valid.
//
// Contoh ekspresi yang valid:
//   "DB_TYPE != 'None'"
//   "USE_REDIS == true"
//   "DB_TYPE == 'PostgreSQL' && USE_MIGRATIONS == true"
func Evaluate(expression string, context map[string]any) (bool, error)
```

### Contoh Perilaku yang Diharapkan

Buat tabel test case berikut sebagai panduan implementasi:

| Ekspresi | Context | Expected |
|---|---|---|
| `"DB_TYPE != 'None'"` | `{"DB_TYPE": "PostgreSQL"}` | `true` |
| `"DB_TYPE != 'None'"` | `{"DB_TYPE": "None"}` | `false` |
| `"USE_REDIS == true"` | `{"USE_REDIS": false}` | `false` |
| `"DB_TYPE == 'PostgreSQL' && USE_REDIS == true"` | `{"DB_TYPE": "PostgreSQL", "USE_REDIS": true}` | `true` |
| `""` (string kosong) | apapun | `true` (tidak ada kondisi = selalu true) |

Ekspresi kosong atau kosong-spasi harus selalu mengembalikan `true` tanpa error — ini digunakan untuk action yang tidak memiliki kondisi `if:`.

---

## Tugas 4 — Context Object (`internal/engine/context.go`)

Context adalah objek pusat yang membawa semua jawaban user dan metadata sepanjang proses scaffolding. Semua bagian engine menerimanya sebagai parameter.

```go
// EngineContext menyimpan semua state yang diperlukan selama satu sesi scaffolding.
type EngineContext struct {
    // Values adalah map dari semua jawaban prompt (key = Prompt.ID, value = jawaban user)
    Values map[string]any

    // Meta berisi informasi tentang template yang sedang digunakan
    Meta BlueprintMeta

    // SourceDir adalah path absolut ke direktori template
    SourceDir string

    // OutputDir adalah path absolut ke direktori output
    OutputDir string

    // DryRun jika true, tidak ada file yang ditulis ke disk
    DryRun bool

    // NoHooks jika true, post-scaffold hooks tidak dijalankan
    NoHooks bool

    // Format adalah output format: "human" atau "json"
    Format string
}

type BlueprintMeta struct {
    Name    string
    Version string
    Source  string
    Commit  string
}

// Get mengambil nilai dari context dengan type assertion yang aman.
// Mengembalikan nilai default jika key tidak ditemukan.
func (c *EngineContext) Get(key string) any

// GetString mengambil nilai sebagai string. Mengembalikan "" jika tidak ada atau bukan string.
func (c *EngineContext) GetString(key string) string

// GetBool mengambil nilai sebagai bool. Mengembalikan false jika tidak ada atau bukan bool.
func (c *EngineContext) GetBool(key string) bool
```

---

## Tugas 5 — Template Renderer (`internal/engine/renderer.go`)

Renderer mengambil file `.tmpl`, memasukkan nilai dari `EngineContext`, dan menghasilkan konten file final.

### Spesifikasi Fungsi

```go
// RenderFile membaca file template dari sourcePath, mengaplikasikan context,
// dan mengembalikan konten yang sudah di-render sebagai string.
// Mendukung file .tmpl (Go template) maupun file biasa (dikopi apa adanya).
func RenderFile(sourcePath string, ctx *EngineContext) (string, error)

// RenderString mengaplikasikan context ke template string yang diberikan langsung.
// Digunakan untuk me-render path target dan nilai-nilai dinamis lainnya.
func RenderString(tmplStr string, ctx *EngineContext) (string, error)
```

### Ketentuan Penting

File yang diakhiri dengan `.tmpl` diproses menggunakan `text/template`. File lain dikopi apa adanya tanpa rendering. Ini penting untuk file binary atau file yang mengandung karakter yang bertabrakan dengan syntax template.

Path target pada setiap action juga harus di-render melalui `RenderString` karena bisa mengandung template variables, misalnya `./internal/{{.PROJECT_NAME}}/handler.go`.

Tambahkan custom template functions yang berguna:

```go
// Fungsi template tambahan yang tersedia di semua file .tmpl:
// - toLower   : mengubah string ke lowercase
// - toUpper   : mengubah string ke uppercase
// - toCamel   : mengubah string ke CamelCase (untuk nama struct)
// - toSnake   : mengubah string ke snake_case (untuk nama file)
// - hasPrefix : mengecek apakah string diawali oleh prefix tertentu
// - now       : mengembalikan waktu saat ini dalam format RFC3339
```

---

## Tugas 6 — File Writer (`internal/engine/writer.go`)

Writer bertanggung jawab untuk menulis konten yang sudah di-render ke file di direktori output. Writer harus mendukung dry-run mode.

### Spesifikasi Fungsi

```go
// WriteFile menulis konten ke targetPath.
// Jika direktori parent belum ada, buat secara rekursif.
// Jika DryRun true pada context, fungsi ini hanya log aksi tanpa menulis ke disk.
// Jika file sudah ada, tanyakan konfirmasi ke user KECUALI flag --yes aktif.
func WriteFile(targetPath string, content string, ctx *EngineContext) error

// WriteResult adalah hasil dari satu operasi write, digunakan untuk laporan akhir
type WriteResult struct {
    Path    string
    Action  string  // "created" | "skipped" | "modified" | "dry-run"
    Error   error
}
```

### Ketentuan Bounded Concurrency

Jika di kemudian hari writer dimodifikasi untuk menulis file secara paralel, gunakan `semaphore` dengan batas maksimal 10 goroutine concurrent untuk menghindari exhausting file descriptors pada proyek besar. Buat komentar pengingat ini di dalam kode.

---

## Tugas 7 — File Tree Walker (`internal/engine/walker.go`)

Walker menyisir semua file di direktori template dan menghasilkan daftar action yang perlu dieksekusi berdasarkan blueprint dan context.

```go
// Walk menyisir direktori template dan mengembalikan daftar ResolvedAction
// yang sudah dievaluasi kondisinya berdasarkan context yang diberikan.
func Walk(blueprint *Blueprint, ctx *EngineContext) ([]ResolvedAction, error)

// ResolvedAction adalah action dari blueprint yang sudah dievaluasi —
// kondisi `if:` sudah diperiksa, path sudah di-render.
type ResolvedAction struct {
    Original   Action
    SourcePath string  // Path absolut ke file sumber
    TargetPath string  // Path absolut ke file tujuan (sudah di-render)
    ShouldRun  bool    // Hasil evaluasi kondisi `if:`
    SkipReason string  // Penjelasan mengapa di-skip (jika ShouldRun == false)
}
```

---

## Tugas 8 — Hook Runner (`internal/engine/hooks.go`)

Hook runner menjalankan perintah shell yang didefinisikan sebagai action dengan `type: "shell"` dalam blueprint.

```go
// RunHook menjalankan satu shell command dalam working directory yang ditentukan.
// Output dari command di-stream ke stdout secara real-time.
// Mengembalikan error jika command keluar dengan exit code non-zero.
func RunHook(action Action, ctx *EngineContext) error
```

Setiap hook harus dijalankan dalam working directory yang benar (`action.WorkingDir` setelah di-render dengan context). Jika `working_dir` kosong, gunakan `ctx.OutputDir` sebagai default. Stdout dan stderr dari hook harus diforward ke terminal user secara real-time, bukan dikumpulkan dan ditampilkan setelah selesai.

---

## Tugas 9 — Lock File Generator (`internal/lock/`)

### `internal/lock/writer.go`

```go
// LockFile merepresentasikan konten dari file symphony.lock
type LockFile struct {
    SymphonyVersion string            `json:"symphony_version"`
    GeneratedAt     time.Time         `json:"generated_at"`
    Template        TemplateLockInfo  `json:"template"`
    Inputs          map[string]any    `json:"inputs"`
    OutputChecksum  string            `json:"output_checksum"`
}

type TemplateLockInfo struct {
    Source  string `json:"source"`
    Version string `json:"version"`
    Commit  string `json:"commit"`
}

// Write menulis LockFile ke direktori output sebagai "symphony.lock"
func Write(lockFile *LockFile, outputDir string) error
```

### `internal/lock/reader.go`

```go
// Read membaca dan mem-parse file symphony.lock dari direktori yang diberikan.
// Mengembalikan error yang deskriptif jika file tidak ditemukan atau corrupt.
func Read(dir string) (*LockFile, error)
```

---

## Tugas 10 — Engine Orchestrator (`internal/engine/engine.go`)

Engine adalah koordinator utama yang menghubungkan semua komponen di atas. Ini adalah implementasi dari workflow tiga fase: Init → Execute → Finalize.

```go
// Engine adalah orchestrator utama scaffolding
type Engine struct {
    blueprint *Blueprint
    ctx       *EngineContext
}

// New membuat instance Engine baru dari blueprint dan context yang diberikan
func New(blueprint *Blueprint, ctx *EngineContext) *Engine

// Run menjalankan proses scaffolding lengkap:
// 1. Validasi blueprint dan context
// 2. Jalankan Walker untuk mendapatkan daftar ResolvedAction
// 3. Tampilkan dry-run summary (selalu, minta konfirmasi jika bukan --yes)
// 4. Jika DryRun, berhenti di sini
// 5. Iterasi setiap ResolvedAction dan render+write file
// 6. Jalankan semua shell hooks
// 7. Generate symphony.lock
// 8. Tampilkan completion message
func (e *Engine) Run() error
```

---

## Tugas 11 — Hubungkan Engine ke Command `gen`

Sekarang hubungkan semua komponen ke command `symphony gen` di `cmd/gen.go`. Command ini masih menggunakan prompt `fmt.Scan` atau `bufio.Scanner` yang sederhana untuk mengumpulkan jawaban user — TUI Bubbletea akan ditambahkan di Fase 2.

Command `gen` harus:
1. Menerima argument pertama sebagai source path template
2. Membaca flag `--out`, `--dry-run`, `--no-hooks`, `--yes`
3. Memanggil `blueprint.Parse()` untuk memuat template
4. Mengumpulkan jawaban prompt via stdin sederhana (loop `fmt.Scan`)
5. Membuat `EngineContext` dari jawaban
6. Membuat `Engine` dan memanggil `Run()`
7. Menangani semua error dengan format yang sesuai (lihat Tugas 12)

---

## Tugas 12 — Error Formatting (`internal/engine/`)

Semua error yang ditampilkan ke user harus mengikuti format yang konsisten. Buat helper function untuk ini:

```go
// PrintError menampilkan error dalam format Symphony yang konsisten:
//
//   ✖ [Judul Error]
//   ─────────────────────────────────────
//   Field:    PROJECT_NAME
//   Masalah:  Deskripsi masalah
//   Saran:    Cara memperbaikinya
//
func PrintError(title, field, problem, suggestion string)
```

Format ini berlaku untuk output "human". Jika `--format json` aktif, print sebagai JSON event:
```json
{"type": "error", "code": 2, "field": "PROJECT_NAME", "message": "...", "suggestion": "..."}
```

---

## Tugas 13 — Unit Tests Wajib

Tulis unit test untuk setiap komponen berikut. Gunakan `testify` untuk assertions.

### `pkg/expr/eval_test.go`
Test semua kombinasi ekspresi di tabel Tugas 3. Tambahkan test untuk ekspresi yang tidak valid (harus mengembalikan error, bukan panic).

### `internal/blueprint/parser_test.go`
Test parsing file YAML yang valid, YAML yang corrupt, file yang tidak ada, dan blueprint yang tidak memiliki field wajib.

### `internal/engine/renderer_test.go`
Test rendering file `.tmpl` dengan berbagai nilai context. Test khusus untuk conditional block (`{{if}}`) yang harus muncul atau tidak muncul berdasarkan context. Test rendering path target yang mengandung template variable.

### `testdata/templates/simple-go/`
Buat template fixture sederhana yang digunakan oleh integration test. Template ini harus memiliki:
- `template.yaml` dengan minimal dua prompt dan satu conditional action
- Dua file `.tmpl` sederhana
- Satu file non-template yang harus dikopi apa adanya

---

## Checklist Selesai Fase 1

- [ ] `symphony gen ./testdata/templates/simple-go --out /tmp/test-output` menghasilkan file dengan benar
- [ ] Conditional actions bekerja: file yang ber-`if:` hanya dibuat jika kondisi terpenuhi
- [ ] `symphony.lock` terbuat di direktori output
- [ ] Post-scaffold hooks dijalankan dan output-nya tampil di terminal
- [ ] Dry-run mode (`--dry-run`) menampilkan preview tanpa membuat file
- [ ] `go test ./pkg/expr/... -v` semua pass
- [ ] `go test ./internal/blueprint/... -v` semua pass
- [ ] `go test ./internal/engine/... -v` semua pass
- [ ] `go build ./...` berjalan tanpa error

---

## Catatan Penting untuk AI

Implementasikan setiap komponen secara independen dan testable. Hindari coupling langsung antar package — gunakan interface di mana memungkinkan, terutama antara `engine` dan `tui` (yang belum ada). Engine tidak boleh mengimport package `tui` secara langsung; gunakan callback atau interface untuk output. Ini akan memudahkan penggantian output mode antara plain text dan TUI di Fase 2 tanpa mengubah logika engine.
