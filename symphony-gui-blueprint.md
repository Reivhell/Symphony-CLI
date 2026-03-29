# Symphony Desktop — GUI Addendum
> *Visual Scaffolding Interface for Symphony CLI*

**Type:** External Blueprint · Addendum to `symphony-cli-blueprint.md`  
**Version:** 0.1.0-draft  
**Status:** Planned · Fase 5 (Post-CLI Stabilization)  
**Prerequisite:** Symphony CLI >= 0.4.0 dengan `--format json` flag tersedia  
**Last Updated:** 2025-03-26

---

## Table of Contents

1. [Konteks & Motivasi](#1-konteks--motivasi)
2. [Tech Stack Decision](#2-tech-stack-decision)
3. [Arsitektur Integrasi CLI ↔ GUI](#3-arsitektur-integrasi-cli--gui)
4. [Project Structure](#4-project-structure)
5. [UI/UX Design System](#5-uiux-design-system)
6. [Screen & Feature Breakdown](#6-screen--feature-breakdown)
7. [State Management](#7-state-management)
8. [CLI Bridge Layer](#8-cli-bridge-layer)
9. [Distribusi & Packaging](#9-distribusi--packaging)
10. [Persiapan yang Diperlukan di Sisi CLI](#10-persiapan-yang-diperlukan-di-sisi-cli)
11. [Roadmap Fase 5](#11-roadmap-fase-5)
12. [Architect's Notes](#12-architects-notes)

---

## 1. Konteks & Motivasi

Symphony CLI dirancang sebagai alat yang hidup di terminal. Namun ada segmen pengguna yang akan mendapat manfaat signifikan dari antarmuka visual — terutama developer yang baru onboarding ke suatu stack, anggota tim yang tidak terbiasa dengan CLI, atau siapapun yang ingin mengeksplorasi template yang tersedia tanpa menghafal syntax perintah.

Symphony Desktop hadir bukan sebagai pengganti CLI, melainkan sebagai **lapisan presentasi alternatif** di atas engine yang sama. Prinsip ini tidak boleh dikompromikan: seluruh logika scaffolding tetap berada di CLI binary, dan GUI hanya bertanggung jawab atas presentasi dan pengumpulan input.

Implikasi langsungnya adalah paritas fitur terjaga secara otomatis — setiap fitur baru yang ditambahkan ke CLI akan tersedia di GUI tanpa memerlukan perubahan pada kode antarmuka, selama kontrak output JSON dipatuhi.

---

## 2. Tech Stack Decision

### 2.1 Framework GUI: Tauri v2

Setelah mempertimbangkan Tauri dan Wails, **Tauri v2 adalah rekomendasi utama** untuk Symphony Desktop. Berikut adalah perbandingan keduanya secara objektif.

| Kriteria | Tauri v2 | Wails v2 |
|---|---|---|
| **Bahasa backend** | Rust | Go (sama dengan Symphony) |
| **Ukuran binary** | ~3–8 MB | ~8–15 MB |
| **Ekosistem & komunitas** | Lebih besar, aktif, didukung penuh | Lebih kecil, berkembang |
| **Keamanan** | CSP ketat, audited | Standar |
| **WebView** | OS WebView (WKWebView / WebView2 / GTK) | OS WebView |
| **Integrasi dengan Go binary** | Via `sidecar` atau `shell` API | Native — Go adalah backend |
| **Kompleksitas build** | Membutuhkan Rust toolchain | Hanya butuh Go |
| **Plugin ekosistem** | Kaya (file dialog, updater, dll) | Terbatas |

Wails memiliki keunggulan nyata dalam simplisitas — backend ditulis dalam Go sehingga tidak perlu belajar Rust. Namun Symphony Desktop perlu dibangun dengan fondasi yang tahan lama. Tauri memiliki komunitas yang lebih besar, dukungan keamanan yang lebih matang, dan ekosistem plugin yang jauh lebih kaya. Kompleksitas Rust di sisi backend Tauri juga minimal karena hampir semua logika dijalankan oleh Symphony CLI binary sebagai sidecar — Rust hanya menjadi jembatan tipis antara frontend dan CLI.

Dengan demikian, **Tauri adalah investasi yang lebih baik untuk jangka panjang**, terutama jika Symphony suatu saat berkembang menjadi open source project dengan banyak kontributor.

### 2.2 Frontend Stack

| Komponen | Teknologi | Versi | Alasan |
|---|---|---|---|
| **Build Tool** | Vite | 5.x | HMR instan, konfigurasi minimal, native ESM |
| **UI Framework** | React | 18.x | Ekosistem terluas, familiar bagi mayoritas kontributor |
| **Language** | TypeScript | 5.x | Type safety untuk kontrak data dengan CLI |
| **Styling** | Tailwind CSS | 3.x | Utility-first, konsisten, mudah dikustomisasi |
| **Component Library** | shadcn/ui | latest | Headless, tidak opinionated, bisa dikustomisasi penuh |
| **State Management** | Zustand | 4.x | Ringan, tidak boilerplate-heavy seperti Redux |
| **Async/Server State** | TanStack Query | 5.x | Untuk mengelola state panggilan ke CLI binary |
| **Routing** | TanStack Router | 1.x | Type-safe routing untuk multi-page layout |
| **Icons** | Lucide React | latest | Konsisten, ringan |
| **Animation** | Framer Motion | 11.x | Transisi antar screen yang halus |
| **Code Highlighting** | Shiki | 1.x | Syntax highlighting untuk file preview pane |

### 2.3 Backend Tauri (Rust Layer)

Rust layer di Tauri sengaja dijaga sesimpel mungkin. Tanggung jawabnya hanya tiga hal: mengelola lifecycle Symphony CLI sidecar, meneruskan perintah dan membaca output-nya, serta mengakses API sistem operasi yang tidak bisa dijangkau dari frontend (dialog file, notifikasi, path resolving).

```toml
# src-tauri/Cargo.toml — dependencies yang relevan
[dependencies]
tauri = { version = "2", features = ["shell-open", "dialog", "notification", "updater"] }
tauri-plugin-shell = "2"      # Untuk menjalankan symphony CLI sidecar
tauri-plugin-dialog = "2"     # File/folder picker dialog
tauri-plugin-fs = "2"         # Akses filesystem (baca symphony.lock, dll)
tauri-plugin-updater = "2"    # Auto-update Symphony Desktop
serde = { version = "1", features = ["derive"] }
serde_json = "1"
```

---

## 3. Arsitektur Integrasi CLI ↔ GUI

### 3.1 Pola Integrasi: Sidecar Binary

Symphony CLI di-bundle sebagai **sidecar binary** di dalam paket Tauri. Ini berarti pengguna menginstal satu aplikasi desktop dan mendapatkan keduanya — GUI dan CLI — sekaligus. CLI tetap dapat digunakan secara mandiri dari terminal setelah diinstal.

```
┌─────────────────────────────────────────────────────────────┐
│                   Symphony Desktop (Tauri)                  │
│                                                             │
│   ┌─────────────────────────┐                               │
│   │   React Frontend        │                               │
│   │   (Vite + TypeScript)   │◄──── Tauri IPC (invoke)      │
│   └─────────────────────────┘                               │
│                  │                                          │
│                  ▼                                          │
│   ┌─────────────────────────┐                               │
│   │   Tauri Rust Backend    │                               │
│   │   (Thin bridge layer)   │                               │
│   └─────────────────────────┘                               │
│                  │                                          │
│      spawn / stdin / stdout / stderr                        │
│                  │                                          │
│                  ▼                                          │
│   ┌─────────────────────────┐                               │
│   │  symphony CLI (sidecar) │  ← binary Go yang sama       │
│   │  --format json          │    persis dengan CLI          │
│   └─────────────────────────┘                               │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Alur Komunikasi

Setiap interaksi GUI mengikuti alur yang deterministik dan satu arah dari sisi inisiasi.

```
User Action (React)
      │
      │  tauri invoke("run_symphony", { args: [...] })
      ▼
Rust Command Handler
      │
      │  sidecar.spawn(["gen", "--format", "json", ...])
      ▼
Symphony CLI binary
      │
      │  stdout stream: newline-delimited JSON events
      ▼
Rust: parse & forward via tauri event
      │
      │  window.emit("symphony:event", payload)
      ▼
React: useEffect listener → update Zustand store → re-render
```

### 3.3 JSON Event Protocol

CLI Symphony harus mengimplementasikan structured JSON output dengan flag `--format json`. Setiap event dikirim sebagai satu baris JSON ke `stdout` (newline-delimited JSON / NDJSON).

```jsonc
// Event: progress
{ "type": "progress", "current": 5, "total": 12, "file": "internal/domain/user/entity.go", "action": "CREATE" }

// Event: file_skip
{ "type": "file_skip", "file": "internal/infrastructure/redis/cache.go", "reason": "USE_REDIS == false" }

// Event: hook_start
{ "type": "hook_start", "command": "go mod tidy" }

// Event: hook_output
{ "type": "hook_output", "line": "go: downloading github.com/lib/pq v1.10.9" }

// Event: complete
{ "type": "complete", "files_created": 12, "files_skipped": 3, "duration_ms": 1842 }

// Event: error
{ "type": "error", "code": 2, "field": "PROJECT_NAME", "message": "...", "suggestion": "..." }

// Event: dry_run_summary (untuk preview pane)
{ "type": "dry_run_summary", "actions": [
    { "action": "CREATE", "path": "cmd/main.go" },
    { "action": "SKIP",   "path": "internal/infrastructure/redis/", "reason": "USE_REDIS == false" }
]}
```

---

## 4. Project Structure

```
symphony-desktop/
│
├── src/                              # React frontend
│   ├── app/
│   │   ├── router.tsx                # TanStack Router setup
│   │   └── providers.tsx             # Zustand + TanStack Query providers
│   │
│   ├── screens/
│   │   ├── Home/                     # Layar utama: template browser
│   │   ├── Wizard/                   # Multi-step scaffolding wizard
│   │   │   ├── StepSource.tsx        # Pilih template source
│   │   │   ├── StepPrompts.tsx       # Form pertanyaan dinamis
│   │   │   ├── StepPreview.tsx       # Dry-run file tree preview
│   │   │   └── StepOutput.tsx        # Pilih output directory
│   │   ├── Generation/               # Live generation progress screen
│   │   └── Settings/                 # Konfigurasi global Symphony
│   │
│   ├── components/
│   │   ├── ui/                       # shadcn/ui base components
│   │   ├── FileTreePreview/          # Komponen file tree dengan warna action
│   │   ├── ProgressTerminal/         # Terminal-style output viewer
│   │   ├── TemplateCard/             # Card untuk template browser
│   │   └── PromptField/              # Dynamic form field (select, input, confirm)
│   │
│   ├── stores/
│   │   ├── wizardStore.ts            # State wizard: answers, current step
│   │   ├── generationStore.ts        # State generation: progress, events
│   │   └── settingsStore.ts          # Global settings
│   │
│   ├── hooks/
│   │   ├── useSymphony.ts            # Hook utama: invoke CLI via Tauri
│   │   ├── useDryRun.ts              # Hook untuk live preview
│   │   └── useTemplateList.ts        # Hook untuk fetch daftar template
│   │
│   ├── types/
│   │   ├── blueprint.ts              # TypeScript interface untuk template.yaml
│   │   ├── events.ts                 # TypeScript interface untuk JSON events
│   │   └── symphony.ts               # Shared types
│   │
│   └── lib/
│       ├── tauri.ts                  # Wrapper untuk Tauri invoke/listen
│       └── utils.ts
│
├── src-tauri/                        # Tauri / Rust backend
│   ├── src/
│   │   ├── main.rs
│   │   ├── commands/
│   │   │   ├── symphony.rs           # run_symphony, get_version, dry_run
│   │   │   ├── filesystem.rs         # resolve_path, open_in_explorer
│   │   │   └── templates.rs          # list_cached_templates, clear_cache
│   │   └── sidecar.rs                # Symphony CLI sidecar manager
│   │
│   ├── binaries/                     # Symphony CLI sidecar binaries
│   │   ├── symphony-x86_64-pc-windows-msvc.exe
│   │   ├── symphony-x86_64-apple-darwin
│   │   ├── symphony-aarch64-apple-darwin
│   │   └── symphony-x86_64-unknown-linux-gnu
│   │
│   ├── icons/                        # App icons (semua ukuran)
│   ├── Cargo.toml
│   └── tauri.conf.json
│
├── public/
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

---

## 5. UI/UX Design System

### 5.1 Design Philosophy

Symphony Desktop mengikuti prinsip **"Terminal-Inspired, Not Terminal-Constrained"** — antarmuka yang terasa familiar bagi developer yang terbiasa dengan terminal, namun menawarkan interaksi yang lebih kaya: drag-and-drop, real-time preview, dan navigasi visual.

Tampilan menggunakan tema gelap secara default dengan pilihan tema terang, mengikuti preferensi sistem operasi. Warna utama diwarisi dari palette Symphony CLI untuk menjaga konsistensi identitas visual antar kedua antarmuka.

### 5.2 Color Palette

| Token | Light | Dark | Kegunaan |
|---|---|---|---|
| `--color-brand` | `#5F87FF` | `#5F87FF` | Aksi utama, tombol primer, progress |
| `--color-success` | `#16A34A` | `#04B575` | File CREATE, operasi sukses |
| `--color-warning` | `#CA8A04` | `#FFD700` | File MODIFY, peringatan |
| `--color-danger` | `#DC2626` | `#FF5F57` | Error, File DELETE |
| `--color-muted` | `#6B7280` | `#626262` | File SKIP, secondary text |
| `--color-surface` | `#F9FAFB` | `#1A1A2E` | Background utama |
| `--color-panel` | `#FFFFFF` | `#16213E` | Panel, card background |
| `--color-border` | `#E5E7EB` | `#2D3748` | Garis batas, divider |

### 5.3 Typography

Seluruh antarmuka menggunakan dua font: **Inter** untuk UI text (label, body, heading) dan **JetBrains Mono** untuk konten yang bersifat kode — path file, terminal output, dan template preview. Keduanya di-bundle bersama aplikasi untuk memastikan konsistensi lintas sistem operasi.

### 5.4 Layout Anatomy

```
┌───────────────────────────────────────────────────────────────┐
│  ◆ Symphony Desktop          [─] [□] [✕]    macOS / Windows  │
├──────────────┬────────────────────────────────────────────────┤
│              │                                                │
│   Sidebar    │                Main Panel                      │
│   (240px)    │                                                │
│              │                                                │
│  Navigation  │   Content area — berubah sesuai screen aktif   │
│  Template    │                                                │
│  History     │                                                │
│              │                                                │
├──────────────┴────────────────────────────────────────────────┤
│  Status Bar: versi Symphony CLI · template aktif · log        │
└───────────────────────────────────────────────────────────────┘
```

---

## 6. Screen & Feature Breakdown

### 6.1 Home Screen — Template Browser

Layar pertama yang dilihat pengguna setelah membuka aplikasi. Menampilkan template yang tersedia dari berbagai sumber dalam format card grid.

Setiap template card menampilkan nama, deskripsi singkat, tags (misalnya `go`, `hexagonal`, `postgresql`), versi, dan jumlah bintang GitHub jika template berasal dari remote. Filter tersedia berdasarkan tag dan bahasa pemrograman. Tombol "Use Template" di setiap card membawa pengguna langsung ke Wizard.

### 6.2 Wizard Screen — Multi-Step Scaffolding

Ini adalah layar inti Symphony Desktop. Wizard dibagi menjadi empat langkah linear dengan progress indicator di bagian atas.

**Langkah 1 — Source.** Pengguna memasukkan sumber template: path lokal (dengan tombol folder picker), URL GitHub/GitLab, atau memilih dari template yang sudah ter-cache. Symphony menjalankan `symphony check --format json` di background untuk memvalidasi template dan menampilkan metadata-nya secara real-time.

**Langkah 2 — Configure.** Form pertanyaan dirender secara dinamis berdasarkan `prompts` yang didefinisikan dalam `template.yaml`. Setiap jenis prompt dirender sebagai komponen yang sesuai — `select` menjadi dropdown, `confirm` menjadi toggle switch, `input` menjadi text field dengan validasi inline. Field dengan `depends_on` otomatis muncul atau menghilang saat jawaban yang relevan berubah, tanpa reload.

**Langkah 3 — Preview.** Setelah semua prompt terisi, Symphony menjalankan `symphony gen --dry-run --format json` dan menampilkan hasilnya sebagai interactive file tree. Setiap file diberi label berwarna sesuai aksinya (CREATE, SKIP, MODIFY). Pengguna dapat mengklik setiap file untuk melihat preview konten yang akan digenerate, dengan syntax highlighting dari Shiki.

**Langkah 4 — Output.** Pengguna memilih direktori output menggunakan native folder picker dialog dari Tauri. Setelah direktori dipilih, tombol "Generate" menjadi aktif.

### 6.3 Generation Screen — Live Progress

Setelah pengguna mengkonfirmasi, layar berpindah ke tampilan generation. Layar ini dibagi dua secara vertikal: sisi kiri menampilkan progress bar dan daftar file yang sedang diproses (update real-time dari JSON event stream), sisi kanan menampilkan terminal-style output viewer yang menampilkan log dari post-scaffold hooks.

Saat generation selesai, muncul completion summary yang menampilkan statistik (jumlah file dibuat, waktu yang diperlukan) dan instruksi selanjutnya dari `completion_message` template, dirender sebagai Markdown dengan syntax highlighting.

### 6.4 Settings Screen

Pengaturan global Symphony Desktop: direktori default untuk output, proxy network untuk remote fetching, manajemen cache (tampilkan ukuran, hapus cache), versi Symphony CLI yang aktif, dan toggle untuk auto-update.

---

## 7. State Management

### 7.1 Wizard Store (Zustand)

```typescript
// src/stores/wizardStore.ts
interface WizardState {
  // Source
  templateSource: string;
  templateMeta: TemplateMeta | null;

  // Prompts
  currentStep: number;
  answers: Record<string, unknown>;
  visiblePrompts: Prompt[];       // Subset dari semua prompts (setelah depends_on dievaluasi)

  // Preview
  dryRunResult: DryRunSummary | null;
  selectedPreviewFile: string | null;

  // Output
  outputDirectory: string;

  // Actions
  setAnswer: (id: string, value: unknown) => void;
  nextStep: () => void;
  prevStep: () => void;
  reset: () => void;
}
```

### 7.2 Generation Store (Zustand)

```typescript
// src/stores/generationStore.ts
interface GenerationState {
  status: 'idle' | 'running' | 'complete' | 'error';
  events: SymphonyEvent[];        // Semua event dari JSON stream
  progress: { current: number; total: number };
  currentFile: string;
  hookOutput: string[];
  completionMessage: string;
  error: SymphonyError | null;
}
```

---

## 8. CLI Bridge Layer

### 8.1 Tauri Command (Rust)

```rust
// src-tauri/src/commands/symphony.rs

#[tauri::command]
async fn run_symphony(
    app: tauri::AppHandle,
    args: Vec<String>,
    window: tauri::Window,
) -> Result<(), String> {
    let sidecar = app
        .shell()
        .sidecar("symphony")
        .map_err(|e| e.to_string())?;

    // Selalu append --format json untuk structured output
    let mut full_args = args;
    full_args.extend(["--format".to_string(), "json".to_string()]);

    let (mut rx, _child) = sidecar
        .args(&full_args)
        .spawn()
        .map_err(|e| e.to_string())?;

    // Stream setiap baris output sebagai Tauri event ke frontend
    while let Some(event) = rx.recv().await {
        if let CommandEvent::Stdout(line) = event {
            let line_str = String::from_utf8_lossy(&line).to_string();
            window
                .emit("symphony:event", &line_str)
                .map_err(|e| e.to_string())?;
        }
    }
    Ok(())
}
```

### 8.2 React Hook

```typescript
// src/hooks/useSymphony.ts
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { useGenerationStore } from "@/stores/generationStore";

export function useSymphony() {
  const addEvent = useGenerationStore((s) => s.addEvent);

  const runGen = async (args: string[]) => {
    // Pasang listener sebelum menjalankan command
    const unlisten = await listen<string>("symphony:event", (event) => {
      const parsed: SymphonyEvent = JSON.parse(event.payload);
      addEvent(parsed);
    });

    try {
      await invoke("run_symphony", { args });
    } finally {
      unlisten();
    }
  };

  return { runGen };
}
```

---

## 9. Distribusi & Packaging

### 9.1 Artefak yang Dihasilkan

Setiap release Symphony menghasilkan dua set artefak yang independen dan terdistribusi secara terpisah.

**CLI binary** (dari `goreleaser`) didistribusikan sebagai file binary tunggal untuk Linux, macOS (Intel + Apple Silicon), dan Windows — sama seperti yang dirancang di blueprint CLI utama.

**Desktop app** (dari Tauri) didistribusikan sebagai installer native per platform: `.dmg` untuk macOS, `.msi` dan `.exe` installer untuk Windows, serta `.AppImage` dan `.deb` untuk Linux. Symphony CLI binary di-bundle sebagai sidecar di dalam paket desktop sehingga pengguna tidak perlu menginstal CLI secara terpisah.

### 9.2 Versioning Strategy

Versi Symphony Desktop mengikuti versi Symphony CLI yang di-bundle di dalamnya, dengan suffix minor untuk patch desktop-only.

```
Symphony CLI   v0.4.0  →  Symphony Desktop  v0.4.0
Symphony CLI   v0.4.0  →  Symphony Desktop  v0.4.1  (bugfix UI only)
Symphony CLI   v0.5.0  →  Symphony Desktop  v0.5.0
```

### 9.3 Auto-Update

Tauri Plugin Updater digunakan untuk mengirimkan update otomatis ke pengguna Desktop. Update manifest di-host di GitHub Releases. CLI binary yang di-bundle juga ikut diperbarui saat update desktop diterima.

### 9.4 Build Pipeline

```yaml
# .github/workflows/release-desktop.yml (ringkasan)
jobs:
  build-cli:
    # Jalankan goreleaser untuk semua platform
    # Letakkan binary hasil build ke src-tauri/binaries/

  build-desktop:
    needs: build-cli
    strategy:
      matrix:
        platform: [macos-latest, windows-latest, ubuntu-22.04]
    steps:
      - uses: tauri-apps/tauri-action@v0
        with:
          tagName: "desktop-v${{ env.VERSION }}"
```

---

## 10. Persiapan yang Diperlukan di Sisi CLI

Ini adalah daftar pekerjaan yang **harus diselesaikan di sisi CLI** sebelum pengembangan GUI dapat dimulai. Pekerjaan ini sebaiknya masuk ke dalam Fase 2 atau Fase 3 roadmap CLI.

**Structured JSON output (`--format json`).** Seluruh output CLI harus dapat diemit sebagai NDJSON yang dapat di-parse oleh mesin. Ini adalah dependency paling kritis. Tanpa ini, GUI tidak bisa melakukan apapun secara programatik.

**`symphony check` command yang stable.** GUI menggunakan command ini untuk memvalidasi template saat source dimasukkan oleh pengguna. Output-nya harus mencakup metadata template (nama, versi, deskripsi, prompts) dalam format JSON.

**`--dry-run` dengan JSON output.** Preview file tree di Wizard Langkah 3 sepenuhnya bergantung pada output dari `symphony gen --dry-run --format json`. Ini harus menghasilkan daftar lengkap aksi yang akan diambil beserta alasan skip jika ada.

**`--yes` flag untuk non-interactive mode.** GUI mengelola konfirmasi sendiri melalui tombol di antarmuka. Ketika CLI dipanggil dari GUI, semua konfirmasi interaktif harus di-bypass menggunakan flag ini.

**Exit code yang konsisten dan terdokumentasi.** Rust backend perlu mengetahui apakah CLI berakhir sukses atau gagal. Exit code yang tidak konsisten akan menyebabkan GUI menampilkan status yang keliru.

---

## 11. Roadmap Fase 5

### Subfase 5.1 — Foundation (4–6 minggu)
Pekerjaan ini tidak akan menghasilkan UI yang fungsional, tetapi membangun fondasi yang benar adalah satu-satunya cara untuk menghindari refactor besar di subfase berikutnya.

- [ ] Setup proyek Tauri + React + Vite + TypeScript + Tailwind
- [ ] Konfigurasi Symphony CLI sebagai sidecar binary
- [ ] Implementasi Rust command `run_symphony` dengan streaming stdout
- [ ] Implementasi React hook `useSymphony` dan Zustand stores
- [ ] Verifikasi JSON event protocol end-to-end (CLI → Rust → React)
- [ ] Setup build pipeline GitHub Actions untuk semua platform

### Subfase 5.2 — Wizard Core (4–6 minggu)
- [ ] Home screen dengan template browser (local + cached)
- [ ] Wizard step 1: template source input + validasi via `symphony check`
- [ ] Wizard step 2: dynamic form rendering dari `prompts` blueprint
- [ ] Wizard step 3: dry-run file tree preview dengan live update
- [ ] Wizard step 4: output directory picker
- [ ] Generation screen dengan real-time progress dan hook output viewer

### Subfase 5.3 — Polish & Distribution (3–4 minggu)
- [ ] Settings screen
- [ ] Remote template browser (search template dari GitHub Topics)
- [ ] Auto-update via Tauri Updater plugin
- [ ] App icon dan splash screen
- [ ] Pengujian di macOS, Windows, dan Linux
- [ ] Dokumentasi instalasi Desktop di website Symphony

---

## 12. Architect's Notes

**Jangan memulai Fase 5 sebelum CLI API stabil.** Setiap perubahan pada struktur JSON event yang dihasilkan CLI akan membutuhkan perubahan di sisi TypeScript types, Rust parsing, dan React store secara bersamaan. Kontrak JSON event harus diperlakukan sebagai public API dan di-freeze sebelum pengembangan GUI dimulai.

**Hindari duplikasi logika validasi.** Validasi input pengguna di sisi GUI (TypeScript) boleh ada untuk memberikan feedback inline yang cepat, tetapi validasi ini harus bersifat superficial — hanya format dasar seperti "apakah field ini kosong". Validasi bisnis yang sesungguhnya tetap menjadi tanggung jawab CLI. Jika keduanya berbeda, CLI adalah otoritas tertinggi.

**Pertimbangkan aksesibilitas dari awal.** shadcn/ui dibangun di atas Radix UI yang sudah memiliki dukungan aksesibilitas (ARIA, keyboard navigation) yang baik. Manfaatkan ini dan jangan menggantinya dengan komponen kustom yang tidak memiliki dukungan serupa.

**Wails sebagai alternatif cepat.** Jika di suatu titik kompleksitas Rust terasa menghambat, migrasi ke Wails adalah pilihan yang valid — terutama karena seluruh backend dapat ditulis dalam Go. Arsitektur sidecar yang dijelaskan di dokumen ini tetap berlaku untuk Wails, hanya lapisan Rust yang digantikan dengan Go.

---

*Dokumen ini adalah addendum dari `symphony-cli-blueprint.md` dan harus dibaca bersama dengan dokumen tersebut.*
