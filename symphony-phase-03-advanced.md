# Symphony CLI — Phase 03: Advanced Features
> **Tipe Dokumen:** AI Build Prompt  
> **Fase:** 3 dari 5  
> **Tujuan:** Implementasi fitur-fitur intelligence — AST Injection, Template Inheritance, Remote Fetching, Caching, Prompt Validation, dan command `check`  
> **Prasyarat:** Phase 02 selesai. TUI berjalan penuh. `symphony gen` dengan local template sudah polished  
> **Output yang diharapkan:** Symphony dapat mengunduh template dari GitHub, template dapat mewarisi template lain, dan engine dapat menyisipkan kode ke file yang sudah ada

---

## Konteks untuk AI

Fase ini menambahkan kemampuan yang membedakan Symphony dari tool-tool sederhana. Tiga pilar utama di fase ini adalah: pertama, kemampuan untuk belajar dari template lain via inheritance; kedua, kemampuan untuk bekerja dengan template yang berada di internet; dan ketiga, kemampuan untuk memodifikasi file kode yang sudah ada, bukan hanya membuat file baru.

Urutan implementasi di fase ini lebih fleksibel dibandingkan fase sebelumnya, namun Remote Fetching harus diselesaikan sebelum Template Inheritance dapat diuji penuh karena `extends:` bisa merujuk ke remote source.

---

## Tugas 1 — Input Validation (`internal/blueprint/validator.go`)

Di Fase 1, validasi blueprint hanya memeriksa field wajib. Di fase ini, implementasikan validasi input yang sebenarnya: memeriksa jawaban user terhadap aturan yang didefinisikan di `validations:` dalam blueprint.

### Spesifikasi

```go
// ValidateInput memvalidasi satu pasang (field, value) terhadap semua aturan
// yang berlaku untuk field tersebut. Mengembalikan slice error — satu per
// aturan yang gagal. Mengembalikan slice kosong jika semua valid.
func ValidateInput(field string, value any, rules []ValidationRule) []ValidationError

type ValidationError struct {
    Field      string
    Rule       string
    Value      any
    Message    string // Dari ValidationRule.Message, atau pesan default jika kosong
    Suggestion string // Contoh nilai yang valid
}
```

### Rules yang Harus Diimplementasikan

**`required`**: Nilai tidak boleh kosong (string kosong, nil, slice kosong). Pesan default: `"Field ini wajib diisi."` Saran default: nama field dalam format yang benar.

**`regex`**: Nilai harus cocok dengan pattern di `ValidationRule.Pattern`. Compile pattern saat startup dan cache hasilnya. Jika pattern tidak valid, kembalikan error konfigurasi (ini adalah bug dalam template, bukan kesalahan user). Saran default: tampilkan pattern dan berikan satu contoh string yang cocok.

**`min_length`**: Untuk string, panjang karakter minimum. `ValidationRule.Pattern` berisi angka sebagai string. Pesan default: `"Minimal N karakter."`.

**`max_length`**: Kebalikan dari `min_length`.

Validasi harus dijalankan secara real-time di komponen TUI input — tampilkan pesan error inline di bawah field, bukan hanya setelah user submit.

---

## Tugas 2 — Prompt Dependency Graph Resolver (`internal/blueprint/resolver.go`)

Di Fase 2, dependency resolution dilakukan secara inline di prompt orchestrator. Ekstrak logika ini ke package `blueprint` agar bisa diuji secara independen.

```go
// ResolveVisible mengembalikan subset dari semua prompts yang harus ditampilkan
// berdasarkan jawaban yang sudah dikumpulkan sejauh ini.
// Urutan prompt dalam slice yang dikembalikan harus dijaga sesuai urutan asli.
func ResolveVisible(prompts []Prompt, collectedAnswers map[string]any) ([]Prompt, error)

// ValidateDependencyGraph memvalidasi bahwa tidak ada circular dependency
// dalam rantai depends_on antar prompts.
// Dipanggil saat blueprint di-parse, bukan saat runtime.
func ValidateDependencyGraph(prompts []Prompt) error
```

Circular dependency harus dideteksi saat parse time, bukan saat runtime. Jika ada circular dependency, `blueprint.Parse()` harus gagal dengan pesan error yang jelas menyebutkan prompt mana yang membentuk siklus.

---

## Tugas 3 — Remote Template Fetching (`internal/remote/fetcher.go`)

Implementasikan kemampuan untuk mengunduh template dari remote source.

### Source Formats yang Didukung

```go
// Fetch mendownload template dari source yang diberikan ke direktori lokal sementara.
// Source bisa berupa:
//   - Path lokal: "./my-template" atau "/home/user/templates/go-api"
//   - GitHub: "github.com/user/repo" atau "github.com/user/repo@v1.2.0"
//   - GitLab: "gitlab.com/user/repo@main"
//   - URL langsung: "https://example.com/template.tar.gz"
//
// Mengembalikan path lokal ke direktori template yang sudah diunduh.
func Fetch(source string, cacheDir string) (localPath string, meta FetchMeta, err error)

type FetchMeta struct {
    Source       string
    ResolvedType string  // "local" | "github" | "gitlab" | "url"
    Version      string  // Tag atau branch
    Commit       string  // Full commit SHA jika tersedia
    CachedAt     time.Time
}
```

### Implementasi untuk Setiap Source Type

**Local path**: Verifikasi direktori ada dan mengandung `template.yaml`. Kembalikan path absolut. Tidak ada caching untuk local path.

**GitHub/GitLab**: Gunakan `go-getter` dari HashiCorp untuk mengunduh. Format URL yang dikonversi ke go-getter: `github.com/user/repo` menjadi `github.com/user/repo//` (double slash adalah syntax go-getter untuk subdirektori). Jika ada `@version`, teruskan sebagai `?ref=version`.

**URL langsung**: Unduh file, deteksi format (zip/tar.gz), ekstrak ke direktori sementara. Verifikasi `template.yaml` ada setelah ekstraksi.

---

## Tugas 4 — Local Cache (`internal/remote/cache.go`)

```go
// CacheDir mengembalikan path ke direktori cache Symphony.
// Default: ~/.symphony/cache/
func CacheDir() string

// CacheKey menghasilkan key yang unik dan aman untuk filesystem
// dari source string yang diberikan.
// Contoh: "github.com/user/repo@v1.2.0" → "github.com_user_repo_v1.2.0"
func CacheKey(source string) string

// IsCached mengecek apakah source sudah ada di cache dan masih valid.
// Template dengan versi tag: selalu valid (immutable).
// Template dengan branch name: valid selama TTL (default 24 jam).
func IsCached(source string, cacheDir string) bool

// CachedPath mengembalikan path ke direktori cache untuk source tertentu.
func CachedPath(source string, cacheDir string) string

// Invalidate menghapus cache untuk source tertentu.
func Invalidate(source string, cacheDir string) error

// List mengembalikan semua entry yang ada di cache beserta metadata-nya.
func List(cacheDir string) ([]CacheEntry, error)

type CacheEntry struct {
    Source    string
    LocalPath string
    CachedAt  time.Time
    SizeBytes int64
    IsTagged  bool  // true jika source menggunakan tag versi (immutable)
}
```

Cache metadata disimpan sebagai file `cache-meta.json` di dalam setiap direktori cache entry. Ini memungkinkan `cache list` menampilkan informasi tanpa harus membaca setiap file template.

---

## Tugas 5 — Template Inheritance (`internal/blueprint/`)

Implementasikan mekanisme `extends:` yang memungkinkan satu template mewarisi dari template lain.

### Aturan Inheritance

**Prompts**: Prompt dari base template datang lebih dulu, diikuti prompt dari child template. Jika ada prompt dengan `id` yang sama di keduanya, prompt dari child template menang (override).

**Actions**: Actions dari base template dieksekusi lebih dulu, diikuti actions dari child template. Child tidak bisa menghapus action dari base.

**Validations**: Digabung (merged). Tidak ada override — semua aturan dari base dan child berlaku.

**CompletionMessage**: Jika child mendefinisikan `completion_message`, gunakan itu. Jika tidak, gunakan dari base.

**Metadata** (`name`, `version`, `author`, dll): Selalu dari child template.

```go
// Resolve mengambil blueprint yang sudah di-parse dan, jika ada field `extends:`,
// mengambil base blueprint (dari local atau remote) dan melakukan merge.
// Mendukung multi-level inheritance (grandparent → parent → child),
// namun harus mendeteksi dan menolak circular inheritance.
func Resolve(child *Blueprint, fetcher Fetcher) (*Blueprint, error)
```

Circular inheritance harus dideteksi. Jika A extends B yang extends A, kembalikan error.

---

## Tugas 6 — AST Injection: Regex-Based (`internal/ast/`)

Implementasikan injeksi kode berbasis anchor comment — pendekatan yang pragmatis dan cukup untuk mayoritas kasus pakai.

### Interface

```go
// internal/ast/injector_interface.go

// Injector adalah interface untuk semua strategi injeksi kode
type Injector interface {
    // Inject membaca file di targetPath, menyisipkan content sesuai strategi,
    // dan menulis hasilnya kembali ke targetPath.
    // Selalu membuat backup sebelum memodifikasi file asli.
    Inject(targetPath string, action Action) error

    // CanHandle mengembalikan true jika injector ini bisa menangani file tersebut
    CanHandle(filePath string) bool
}
```

### Anchor-Based Injector

```go
// internal/ast/anchor_injector.go

// AnchorInjector menyisipkan kode di sekitar anchor comment yang didefinisikan
// dalam template action.
//
// Strategi yang didukung:
//   "after-anchor"  : sisipkan setelah baris yang mengandung anchor
//   "before-anchor" : sisipkan sebelum baris yang mengandung anchor
//   "replace-anchor": ganti baris anchor dengan content
type AnchorInjector struct{}
```

Sebelum melakukan injeksi, selalu:
1. Verifikasi file target ada
2. Buat backup di `<targetPath>.symphony-bak`
3. Cari anchor string — jika tidak ditemukan, kembalikan error yang jelas
4. Lakukan injeksi
5. Jika target adalah file Go, jalankan `gofmt` pada hasil injeksi. Jika `gofmt` gagal, restore backup dan kembalikan error.
6. Hapus backup hanya jika semua langkah berhasil

### `internal/ast/go_injector.go`

Untuk file Go, tambahkan layer tambahan menggunakan `go/parser` dari stdlib untuk memverifikasi bahwa file tetap valid Go setelah injeksi, sebelum menulis ke disk.

---

## Tugas 7 — Command `check` (`cmd/check.go`)

Implementasikan command `symphony check <source>` yang memvalidasi template tanpa menjalankan scaffolding.

Validasi yang harus dilakukan secara berurutan:

Pertama, unduh atau temukan template dari source (gunakan remote fetcher).

Kedua, parse `template.yaml` — laporkan setiap parse error dengan nomor baris.

Ketiga, validasi schema: pastikan semua field wajib ada, tipe data benar, versi kompatibel.

Keempat, validasi dependency graph prompt: tidak ada circular dependency, semua prompt yang direferensikan dalam `depends_on` sudah terdefinisi.

Kelima, validasi actions: semua file source yang direferensikan dalam actions benar-benar ada di direktori template. Jika ada file yang tidak ditemukan, laporkan sebagai warning (bukan error) karena bisa jadi file tersebut di-generate oleh action lain.

Keenam, tampilkan summary hasil validasi.

Output `check` dalam format human:
```
  ◆ Symphony Check
  ──────────────────────────────────────────────────
  Template : Clean-Hexagonal-Go
  Versi    : 1.2.0
  Author   : username

  ✔ Schema valid (schema_version: 2)
  ✔ 7 prompts terdeteksi, dependency graph OK
  ✔ 12 actions terdeteksi
  ✔ Semua file source ditemukan
  ⚠ min_symphony_version: 0.3.0 (kamu menggunakan 0.2.1 — bisa ada incompatibility)

  Hasil: LULUS dengan 1 peringatan
```

---

## Tugas 8 — Command `re-gen` (`cmd/regen.go`)

Implementasikan `symphony re-gen` yang mengulangi scaffolding dari lock file yang ada.

Alur:
1. Cari `symphony.lock` di direktori saat ini — error jika tidak ditemukan
2. Baca lock file menggunakan `lock.Read()`
3. Unduh template dengan versi dan commit yang sama (dari `template.commit`)
4. Buat `EngineContext` dari `inputs` di lock file — tidak ada prompts interaktif
5. Tampilkan dry-run preview
6. Minta konfirmasi — karena ini akan menimpa file yang sudah ada
7. Jalankan engine

---

## Tugas 9 — Command `list` (`cmd/list.go`)

Implementasikan `symphony list` yang menampilkan semua template yang tersedia.

```
  ◆ Symphony Templates
  ──────────────────────────────────────────────────

  Local Templates
  ──────────────────────────────────────────────────
  Tidak ada template lokal yang dikonfigurasi.
  Tambahkan path template di ~/.symphony/config.yaml

  Cached Templates (3)
  ──────────────────────────────────────────────────
  📦 github.com/user/go-blueprint         v1.2.0   3.2 MB   2 hari lalu
  📦 github.com/user/ts-clean-arch        v0.9.1   1.8 MB   5 hari lalu
  📦 github.com/other/rust-actix-web      v2.0.0   4.1 MB   1 minggu lalu

  Jalankan 'symphony cache clear' untuk menghapus semua cache (9.1 MB)
```

---

## Tugas 10 — Update Global Config (`internal/config/config.go`)

Tambahkan field-field yang dibutuhkan oleh fitur baru ke dalam struktur konfigurasi global.

```go
type Config struct {
    // Template paths lokal yang selalu tersedia
    LocalTemplatePaths []string `yaml:"local_template_paths"`

    // TTL untuk cache template dari branch (default: 24 jam)
    CacheTTLHours int `yaml:"cache_ttl_hours"`

    // Proxy untuk networking (opsional)
    HTTPProxy  string `yaml:"http_proxy"`
    HTTPSProxy string `yaml:"https_proxy"`

    // GitHub token untuk menghindari rate limiting saat fetching
    GitHubToken string `yaml:"github_token"`

    // Default output format
    DefaultFormat string `yaml:"default_format"`
}
```

Config dibaca dari `~/.symphony/config.yaml`. Setiap field harus memiliki nilai default yang sane jika file config tidak ada.

---

## Tugas 11 — Integration Tests untuk Fitur Baru

### `internal/remote/fetcher_test.go`
Test fetching dari local path yang valid dan invalid. Test untuk format source yang tidak dikenal. Untuk GitHub fetching, gunakan `httptest.NewServer` untuk mock server — jangan melakukan network call nyata dalam unit test.

### `internal/blueprint/resolver_test.go`
Test inheritance dua level dengan mock blueprint. Test deteksi circular inheritance. Test override prompt dan merge actions.

### `internal/ast/anchor_injector_test.go`
Test injeksi `after-anchor` pada file Go yang valid. Test bahwa backup dibuat dan direstore jika `gofmt` gagal. Test error ketika anchor tidak ditemukan.

---

## Checklist Selesai Fase 3

- [ ] `symphony gen github.com/user/repo --out ./output` mengunduh dan menggunakan template remote
- [ ] Cache dibuat di `~/.symphony/cache/` dan digunakan untuk panggilan berikutnya
- [ ] Template dengan `extends:` berhasil merge prompts dan actions dari base template
- [ ] `symphony check ./my-template` melaporkan validasi dengan benar
- [ ] `symphony re-gen` di direktori dengan `symphony.lock` berhasil mereproduksi scaffold
- [ ] `symphony list` menampilkan cached templates
- [ ] AST injection dengan anchor comment bekerja pada file Go sederhana
- [ ] Circular dependency di prompt `depends_on` terdeteksi saat parse time
- [ ] Circular inheritance di `extends:` terdeteksi dan dilaporkan dengan jelas
- [ ] Semua integration test pass

---

## Catatan Penting untuk AI

Untuk AST injection, hindari ambisi yang terlalu besar di fase ini. Anchor-based injection (cari string, sisipkan di sekitarnya) sudah cukup untuk 95% use case. Implementasi `go/ast` penuh yang bisa mencari "function declaration bernama X" adalah nice-to-have, bukan requirement di fase ini. Prioritaskan correctness dan safety (backup, verifikasi `gofmt`) di atas fitur.

Untuk remote fetching, implementasikan timeout yang sensible (default 30 detik) dan berikan pesan error yang jelas ketika network tidak tersedia — jangan biarkan user menunggu tanpa feedback.
