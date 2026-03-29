# Symphony CLI — Phase 02: Terminal UI/UX
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 2 dari 5  
> **Tujuan:** Mengganti semua interaksi stdin sederhana dengan Terminal User Interface (TUI) yang polished menggunakan Bubbletea, Bubbles, Lipgloss, dan Glamour  
> **Prasyarat:** Phase 01 selesai. `symphony gen` berjalan fungsional dengan prompt stdin sederhana  
> **Output yang diharapkan:** Seluruh alur interaksi Symphony — dari discovery hingga completion — menggunakan TUI interaktif dengan warna, animasi, dan keyboard navigation

---

## Konteks untuk AI

Di fase ini kamu tidak mengubah logika engine sama sekali. Semua yang kamu bangun di Fase 1 tetap dipertahankan. Yang kamu lakukan adalah **mengganti lapisan presentasi**: dari `fmt.Scan` biasa ke komponen TUI yang interaktif.

Arsitektur yang harus diikuti adalah Elm Architecture yang digunakan Bubbletea: setiap komponen TUI memiliki `Model` (state), `Update` (reducer), dan `View` (renderer). Komponen-komponen ini kemudian dikomposisikan oleh `prompt.go` sebagai orchestrator.

Prinsip utama yang harus dijaga: **engine tidak boleh tahu bahwa TUI ada**. Engine menerima `EngineContext` yang sudah terisi — bagaimana context itu diisi (via TUI, stdin, atau file config) bukan urusan engine.

---

## Tugas 1 — Design Tokens dengan Lipgloss

Sebelum membangun komponen apapun, definisikan design tokens terpusat di satu file. Semua komponen TUI harus mengimpor dari file ini — tidak ada hardcoded color di komponen individual.

### `internal/tui/styles.go`

```go
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette Symphony
var (
    ColorBrand    = lipgloss.Color("#5F87FF")
    ColorSuccess  = lipgloss.Color("#04B575")
    ColorWarning  = lipgloss.Color("#FFD700")
    ColorDanger   = lipgloss.Color("#FF5F57")
    ColorMuted    = lipgloss.Color("#626262")
    ColorHighlight = lipgloss.Color("#D787FF")
)

// Base styles — semua komponen derive dari sini
var (
    StyleBase = lipgloss.NewStyle()

    StyleHeader = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorBrand).
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true).
        BorderForeground(ColorMuted).
        Width(54).
        Padding(0, 1)

    StyleDivider = lipgloss.NewStyle().
        Foreground(ColorMuted)

    StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
    StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
    StyleDanger  = lipgloss.NewStyle().Foreground(ColorDanger)
    StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
    StyleBrand   = lipgloss.NewStyle().Foreground(ColorBrand)

    // Label aksi file pada preview
    StyleActionCreate = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
    StyleActionModify = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
    StyleActionSkip   = lipgloss.NewStyle().Foreground(ColorMuted)
    StyleActionDelete = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
)

// Divider — garis pemisah horizontal
func Divider(width int) string {
    return StyleDivider.Render(strings.Repeat("─", width))
}

// Icon helpers
const (
    IconSuccess  = "✔"
    IconError    = "✖"
    IconWarning  = "⚠"
    IconArrow    = "❯"
    IconDiamond  = "◆"
    IconSpinner  = "⠸"
)
```

---

## Tugas 2 — Komponen: Header Banner

Header ditampilkan sekali di awal setiap command yang interaktif.

### `internal/tui/header.go`

```
Output yang diharapkan:

  ◆ Symphony                                    v0.4.0
  ──────────────────────────────────────────────────────
  The Adaptive Scaffolding Engine
```

Implementasikan sebagai fungsi sederhana yang mengembalikan string, bukan sebagai Bubbletea model. Header adalah static UI — tidak ada state yang perlu dikelola.

```go
// RenderHeader mengembalikan string banner Symphony yang siap ditampilkan.
// version adalah string versi CLI yang diteruskan dari cmd layer.
func RenderHeader(version string) string
```

---

## Tugas 3 — Komponen: Text Input

### `internal/tui/input.go`

Komponen untuk prompt bertipe `"input"`. Wrap komponen `textinput` dari library Bubbles dengan behavior Symphony-specific:

Implementasikan sebagai Bubbletea model yang mengikuti pola `Model/Update/View`. Model harus menyimpan: pertanyaan yang ditampilkan, nilai saat ini, placeholder (dari `Prompt.Default`), dan state apakah input sudah disubmit.

Tampilan yang diharapkan:
```
  ? Nama proyek:
  ❯ my-awesome-api
```

Saat user menekan Enter, model masuk ke state `submitted` dan nilai final tersimpan. Komponen harus handle Ctrl+C untuk mengirim sinyal abort ke parent model.

---

## Tugas 4 — Komponen: Single Select

### `internal/tui/select.go`

Komponen untuk prompt bertipe `"select"`. Tampilkan daftar pilihan dengan navigasi keyboard.

Tampilan yang diharapkan:
```
  ? Database apa yang akan digunakan?
  ❯ ● PostgreSQL
    ○ MongoDB
    ○ MySQL
    ○ None
```

Keyboard controls: `↑`/`↓` untuk navigasi, `Enter` untuk memilih, `Ctrl+C` untuk abort. Pilihan yang sedang di-hover ditampilkan dengan warna `ColorHighlight`. Pilihan yang terpilih menggunakan ikon `●`, yang tidak terpilih menggunakan `○`.

---

## Tugas 5 — Komponen: Multi-select

### `internal/tui/multiselect.go`

Komponen untuk prompt bertipe `"multiselect"`. Mirip dengan select tetapi user bisa memilih lebih dari satu opsi.

Tampilan yang diharapkan:
```
  ? Fitur tambahan yang ingin diaktifkan? (Space untuk toggle)
    ☑ Redis Caching
  ❯ ☐ OpenTelemetry Tracing
    ☑ Prometheus Metrics
    ☐ Swagger/OpenAPI
```

Keyboard controls: `↑`/`↓` untuk navigasi, `Space` untuk toggle pilihan, `Enter` untuk konfirmasi semua pilihan, `Ctrl+C` untuk abort. Mengembalikan `[]string` berisi semua opsi yang dipilih.

---

## Tugas 6 — Komponen: Confirm

### `internal/tui/confirm.go`

Komponen untuk prompt bertipe `"confirm"`. Tampilkan pertanyaan yes/no.

Tampilan saat belum dijawab:
```
  ? Gunakan Redis untuk caching? (y/N)
```

Default value dari `Prompt.Default` menentukan huruf kapital: jika default `true`, tampilkan `(Y/n)`, jika `false` tampilkan `(y/N)`. Menekan Enter tanpa input menggunakan default value. Hanya menerima `y`, `Y`, `n`, `N` sebagai input valid.

---

## Tugas 7 — Komponen: Progress Bar

### `internal/tui/progress.go`

Komponen untuk menampilkan progress generation secara real-time.

Tampilan yang diharapkan:
```
  ◆ Generating...
  ──────────────────────────────────────────────────────

  ✔ cmd/main.go
  ✔ internal/domain/user/entity.go
  ⠸ internal/infrastructure/postgres/db.go

  Progress  ████████████░░░░░░░░  8/12 files
```

Gunakan komponen `progress` dari Bubbles sebagai base untuk progress bar. File yang berhasil dibuat langsung muncul di daftar dengan ikon `✔` berwarna `ColorSuccess`. File yang sedang diproses menggunakan spinner `⠸` yang beranimasi.

Komponen ini harus bisa menerima update dari luar (dari engine) via message channel Bubbletea. Gunakan `tea.Cmd` untuk melakukan polling atau gunakan channel message.

---

## Tugas 8 — Komponen: Dry-Run Summary

### `internal/tui/summary.go`

Komponen untuk menampilkan preview sebelum eksekusi dimulai. Ini bukan komponen interaktif penuh — ini adalah tampilan statis yang menunggu konfirmasi.

Tampilan yang diharapkan:
```
  ◆ Review — File yang akan dibuat:
  ──────────────────────────────────────────────────────

  [CREATE]  cmd/main.go
  [CREATE]  internal/domain/user/entity.go
  [CREATE]  internal/usecase/user/service.go
  [CREATE]  internal/infrastructure/postgres/db.go
  [MODIFY]  go.mod                              (+3 deps)
  [SKIP]    internal/infrastructure/redis/      USE_REDIS == false

  Total: 12 file akan dibuat, 1 file dimodifikasi, 3 file di-skip

  ❯ Lanjutkan? (Y/n)
```

Setiap baris file harus diberi warna sesuai aksinya menggunakan style yang sudah didefinisikan di `styles.go`. Setelah konfirmasi, komponen mengembalikan boolean ke parent.

---

## Tugas 9 — Komponen: Completion Screen

### `internal/tui/completion.go`

Layar terakhir yang ditampilkan setelah scaffolding selesai. Render `completion_message` dari blueprint sebagai Markdown menggunakan Glamour.

Tampilan yang diharapkan:
```
  ◆ Scaffold Selesai!
  ──────────────────────────────────────────────────────

  ✔ 12 file dibuat  ✔ go mod tidy selesai  ✔ git init selesai
  ✔ symphony.lock dibuat

  ──────────────────────────────────────────────────────

  [Markdown rendered completion_message dari template]

  ──────────────────────────────────────────────────────
  ✨ Happy coding!
```

Untuk rendering Markdown, gunakan:
```go
renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(72),
)
rendered, _ := renderer.Render(blueprint.CompletionMessage)
```

---

## Tugas 10 — Prompt Orchestrator

### `internal/tui/prompt.go`

Ini adalah komponen paling kompleks di fase ini. Orchestrator mengelola alur seluruh sesi prompting: menampilkan prompt satu per satu, mengevaluasi `depends_on`, mengumpulkan jawaban, dan mengembalikan `EngineContext` yang sudah terisi.

```go
// RunPrompts menjalankan sesi interaktif lengkap untuk semua prompt dalam blueprint.
// Mengembalikan map jawaban yang siap dimasukkan ke EngineContext.
// Mengembalikan error jika user membatalkan (Ctrl+C).
func RunPrompts(blueprint *Blueprint) (map[string]any, error)
```

### Logika Dependency Graph

Sebelum menampilkan setiap prompt, evaluasi `depends_on`-nya menggunakan `pkg/expr`. Jika kondisi tidak terpenuhi berdasarkan jawaban yang sudah dikumpulkan sejauh ini, lewati prompt tersebut tanpa menampilkannya ke user. Gunakan `Prompt.Default` sebagai nilai untuk prompt yang di-skip.

Contoh alur:
```
Prompt 1: DB_TYPE → user jawab "None"
Prompt 2: DB_HOST (depends_on: "DB_TYPE != 'None'") → SKIP (tidak ditampilkan)
Prompt 3: USE_MIGRATIONS (depends_on: "DB_TYPE == 'PostgreSQL'") → SKIP
Prompt 4: USE_REDIS → tampilkan
```

### Composable Model

Orchestrator tidak menjalankan semua prompt dalam satu Bubbletea program. Sebaliknya, jalankan setiap prompt sebagai program Bubbletea yang terpisah secara sekuensial. Ini lebih sederhana dan menghindari kompleksitas parent-child model yang dalam.

```go
for _, prompt := range visiblePrompts {
    answer, err := runSinglePrompt(prompt, collectedAnswers)
    if err != nil {
        return nil, ErrUserCancelled
    }
    collectedAnswers[prompt.ID] = answer
    // Re-evaluasi visible prompts untuk prompt berikutnya
}
```

---

## Tugas 11 — Sambungkan TUI ke Engine via Interface

Buat interface `Reporter` yang memungkinkan engine melaporkan progress tanpa bergantung langsung pada implementasi TUI.

```go
// internal/engine/reporter.go

// Reporter adalah interface yang digunakan engine untuk melaporkan
// progress dan status ke lapisan presentasi (TUI atau JSON output).
type Reporter interface {
    // OnFileCreated dipanggil setiap kali satu file berhasil ditulis
    OnFileCreated(path string)

    // OnFileSkipped dipanggil setiap kali satu file di-skip
    OnFileSkipped(path string, reason string)

    // OnHookStart dipanggil sebelum hook dijalankan
    OnHookStart(command string)

    // OnHookOutput dipanggil untuk setiap baris output dari hook
    OnHookOutput(line string)

    // OnComplete dipanggil ketika seluruh proses selesai
    OnComplete(stats CompletionStats)

    // OnError dipanggil ketika terjadi error
    OnError(err SymphonyError)
}

type CompletionStats struct {
    FilesCreated  int
    FilesSkipped  int
    FilesModified int
    DurationMs    int64
}
```

Buat dua implementasi:
1. `TUIReporter` — mengupdate komponen progress Bubbletea
2. `JSONReporter` — mengoutput event sebagai NDJSON ke stdout (untuk `--format json`)
3. `PlainReporter` — output sederhana `fmt.Println` (fallback)

---

## Tugas 12 — Implementasi `--format json` di Semua Reporter

Setiap event yang dilaporkan `JSONReporter` harus mematuhi protokol berikut. Output ke `stdout`, satu JSON object per baris.

```json
{"type":"progress","current":5,"total":12,"file":"internal/domain/user/entity.go","action":"CREATE"}
{"type":"file_skip","file":"internal/infrastructure/redis/cache.go","reason":"USE_REDIS == false"}
{"type":"hook_start","command":"go mod tidy"}
{"type":"hook_output","line":"go: downloading github.com/lib/pq v1.10.9"}
{"type":"complete","files_created":12,"files_skipped":3,"duration_ms":1842}
{"type":"error","code":2,"field":"PROJECT_NAME","message":"...","suggestion":"..."}
{"type":"dry_run_summary","actions":[{"action":"CREATE","path":"cmd/main.go"},{"action":"SKIP","path":"internal/infrastructure/redis/","reason":"USE_REDIS == false"}]}
```

---

## Tugas 13 — Update `cmd/gen.go`

Sekarang sambungkan semua komponen TUI ke command `gen`. Alur yang harus diimplementasikan:

1. Tampilkan `RenderHeader(version)`
2. Jalankan `blueprint.Parse(source)` — tampilkan spinner selama proses ini
3. Tampilkan hasil discovery (versi template, kompatibilitas)
4. Jalankan `tui.RunPrompts(blueprint)` untuk mengumpulkan jawaban
5. Buat `EngineContext` dari jawaban
6. Jalankan dry-run via engine dan tampilkan `tui.RenderSummary()`
7. Minta konfirmasi kecuali `--yes` aktif
8. Jalankan engine dengan `TUIReporter` (atau `JSONReporter` jika `--format json`)
9. Tampilkan completion screen

---

## Tugas 14 — Keyboard Interrupt Handler

Tangani Ctrl+C di seluruh aplikasi dengan bersih. Saat user menekan Ctrl+C:
- Jangan tampilkan Go's default signal output
- Tampilkan pesan singkat: `\n  Dibatalkan.` dengan warna `ColorMuted`
- Keluar dengan exit code `4`

```go
// Di cmd/root.go — tambahkan signal handler
signal.NotifyContext(ctx, os.Interrupt)
```

---

## Checklist Selesai Fase 2

- [ ] Header banner tampil dengan benar saat `symphony gen` dijalankan
- [ ] Prompt input, select, confirm tampil dengan warna dan styling yang benar
- [ ] `depends_on` bekerja: prompt yang tidak relevan tidak muncul
- [ ] Progress bar update secara real-time saat file digenerate
- [ ] Dry-run summary menampilkan file CREATE/SKIP/MODIFY dengan warna yang tepat
- [ ] Completion screen merender `completion_message` sebagai Markdown
- [ ] Ctrl+C menghasilkan exit code 4 dengan pesan yang bersih
- [ ] `--format json` menghasilkan NDJSON yang valid dan bisa di-parse
- [ ] `--yes` flag melewati semua konfirmasi
- [ ] `go test ./internal/tui/... -v` semua pass

---

## Catatan Penting untuk AI

Jangan overcomplicate model Bubbletea. Untuk setiap komponen prompt, model yang paling sederhana adalah yang terbaik: satu state field `value`, satu state field `done`, dan minimal keyboard handling. Kompleksitas parent-child yang dalam seringkali tidak diperlukan untuk use case ini dan justru membuat debugging lebih sulit. Prioritaskan correctness dan readability di atas feature richness visual.
