<div align="center">

# 🎻 Symphony CLI
**The Adaptive Scaffolding Engine**

[![CI](https://github.com/Reivhell/Symphony-CLI/actions/workflows/ci.yml/badge.svg)](https://github.com/Reivhell/Symphony-CLI/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Reivhell/symphony)](https://goreportcard.com/report/github.com/Reivhell/symphony)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-0.1.0--draft-blue.svg)]()
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg)]()

[English](#-english) | [Bahasa Indonesia](#-bahasa-indonesia)

</div>

---

<br>

## 🇬🇧 English

Symphony is a **data-driven scaffolding orchestrator** that understands the relationship between components in software architecture. It reads dynamic `template.yaml` blueprints, adapts outputs based on user context, and automates repetitive development rituals.

[Installation](#-installation) • [Quick Start](#-quick-start) • [How It Works](#-how-it-works) • [Command Reference](#-command-reference) • [Documentation](#-documentation)

### 🌟 Why Symphony?

Unlike traditional code generators like *Cookiecutter* (which can be too basic) or *Nx/Turbo* (which are highly opinionated), **Symphony** sits in the sweet spot for modern polyglot environments:

- **🧠 Zero Assumption Core:** The CLI engine is completely decoupled from the templates. It knows nothing about the languages or frameworks being generated.
- **⚡ Conditional Scaffolding:** Skip files and directories dynamically based on logic rules encoded in the blueprint (e.g., `if: "DB_TYPE == 'PostgreSQL'"`).
- **🏗️ Composable by Design:** Templates can inherit (`extend`) other base templates, eliminating boilerplate duplication across your organization's tech stacks.
- **💉 AST Code Injection:** Symphony doesn't just create new files; it can semantically inject code (like new API routes or module imports) into **existing** files safely.
- **🔁 100% Reproducible:** A `symphony.lock` file captures all inputs and exact template versions used, guaranteeing identical scaffolding results across multiple machines or CI/CD pipelines.
- **🛡️ Fail Loudly & Safely:** Before writing a single byte to disk, Symphony presents a dry-run diff preview to validate what will change exactly.

### 🚀 Installation

**Linux / macOS (via Shell Script)**
```bash
curl -sSL https://raw.githubusercontent.com/Reivhell/Symphony-CLI/main/install.sh | sh
```

**Windows**
Download the latest `symphony_*.zip` from [GitHub Releases](https://github.com/Reivhell/Symphony-CLI/releases), extract `symphony.exe`, and add it to your system `PATH`.

**Build from Source (Requires Go 1.22+)**
```bash
go install github.com/Reivhell/symphony@latest
```

### 💻 Quick Start

Generate a new project from a local or remote template seamlessly.

**1. Scaffold from a remote GitHub repository:**
```bash
symphony gen github.com/yourname/go-hexagonal-template@v1.2.0 --out ./my-api
```

**2. Preview changes without writing to disk (Dry-Run Preview):**
```bash
symphony gen github.com/yourname/go-blueprint --out ./my-api --dry-run
```

**3. Reproduce a previous environment precisely from a lock file:**
```bash
# Navigate to a previously generated project directory
symphony re-gen 
```

**4. Scaffold forcefully without interactive prompts (for CI/CD pipelines):**
```bash
symphony gen ./templates/base --out ./app --yes
```

### 🧠 How It Works

Symphony delivers a refined Terminal UX, following a 5-stage progressive workflow:

1. **Discovery & Validation:** Resolves templates from GitHub, GitLab, or local paths. It ensures the `template.yaml` schema is valid and checks minimum toolchain versions.
2. **Interactive Prompting:** Presents dynamic, dependency-aware questions. (For example, the *Database Host* question only appears if a Database was selected in a previous prompt).
3. **Pre-Execution Diff:** Displays a complete, color-coded summary indicating which files will be `[CREATE]`, `[MODIFY]`, or smartly `[SKIP]`.
4. **Generation & Injection:** Compiles `text/template` files, executes abstract syntax tree (AST)/regex node injections, and writes the results to disk concurrently.
5. **Finalization:** Executes arbitrary post-scaffold shell hooks (e.g., `git init`, `go mod tidy`) and generates the final `symphony.lock` file.

### ⌨️ Command Reference

| Command | Description |
|---|---|
| `symphony gen <source>` | Scaffold a new project from a local path, URL, or Git repository. |
| `symphony re-gen` | Repeat a scaffolding process exactly via the `symphony.lock` file. |
| `symphony check <source>` | Validate a template blueprint's health and schema without executing it. |
| `symphony list` | List locally cached or available templates. |
| `symphony update <source>` | Force update a remote template to its latest version in the local cache. |
| `symphony cache clear` | Clear the local template cache index. |

**Global Flags:**
- `--out` / `-o`: Output directory.
- `--dry-run` / `-d`: Preview changes safely into your terminal console.
- `--yes` / `-y`: Skip all interactive confirmations.
- `--config` / `-c`: Provide a custom `.symphony/config.yaml` configuration.

### 📚 Documentation

The heart of Symphony is the `template.yaml` file. Dive deeper into the architecture and configuration:
- [Blueprint/Template Specification](docs/blueprint-spec.md)
- [How to Contribute](docs/contributing.md)
- [Release QC Report](docs/QC_REPORT.md)
- [Changelog](CHANGELOG.md)

### 📄 License

Symphony is open-source software licensed under the [LICENSE](LICENSE) file.

---
<br>

<a name="bahasa-indonesia"></a>

## 🇮🇩 Bahasa Indonesia

Symphony adalah sebuah **data-driven scaffolding orchestrator** yang memahami secara mendalam hubungan antar komponen dalam arsitektur software. Symphony membaca *blueprint* dinamis (`template.yaml`), menyesuaikan *output* berdasarkan konteks pengguna, dan mengotomatisasi ritual-ritual *development* yang repetitif.

[Instalasi](#-instalasi) • [Quick Start](#-quick-start-1) • [Cara Kerja](#-cara-kerja) • [Referensi Command](#-referensi-command) • [Dokumentasi](#-dokumentasi-1)

### 🌟 Mengapa Symphony?

Berbeda dengan *code generator* tradisional seperti *Cookiecutter* (yang terlalu algoritmik-sederhana) atau *Nx/Turbo* (yang terlalu *opinionated*), **Symphony** memegang kendali yang pas untuk ekosistem perangkat lunak moderen:

- **🧠 Zero Assumption Core:** *Engine* CLI ini sepenuhnya terpisah dari template. CLI tidak tahu dan tidak berasumsi tentang bahasa pemrograman atau framework apa yang sedang di-*generate*.
- **⚡ Conditional Scaffolding:** Secara dinamis mengabaikan (*skip*) pembuatan file atau direktori sesuai dengan aturan logika yang ditetapkan di template (contoh: `if: "DB_TYPE == 'PostgreSQL'"`).
- **🏗️ Composable by Design:** Template dapat mewarisi (`extend`) template dasar lainnya, secara drastis menekan duplikasi *boilerplate* di berbagai tech stack organisasi Anda.
- **💉 AST Code Injection:** Symphony tidak hanya membuat file baru; ia bisa menyuntikkan (*inject*) baris kode (misal: penambahan *route* API atau *import* modul baru) secara semantik dan aman langsung berbaur pada file yang **sudah ada**.
- **🔁 100% Reproducible:** File `symphony.lock` dengan tangguh merekam semua jenis input serta versi template yang persis digunakan. Ini menjamin proses *scaffolding* yang identik ke depan pada banyak mesin dan proses CI/CD.
- **🛡️ Fail Loudly & Safely:** Sebelum menulis satu *byte* pun ke *disk*, mode `Dry-Run` pada Symphony akan menyajikan tabel *preview diff* khusus untuk validasi penuh sebelum terjadinya eskalasi kode (perubahan status sistem).

### 🚀 Instalasi

**Linux / macOS (via Shell Script)**
```bash
curl -sSL https://raw.githubusercontent.com/Reivhell/Symphony-CLI/main/install.sh | sh
```

**Windows**
Unduh berkas *zipper* terbaru `symphony_*.zip` melalui [Rilis GitHub Resmi](https://github.com/Reivhell/Symphony-CLI/releases), un-ekstrak `symphony.exe`, lalu sertakan *path* foldernya pada variable `PATH` sistem operasi Anda.

**Kompilasi Manual dari Source (Mewajibkan Go 1.22+)**
```bash
go install github.com/Reivhell/symphony@latest
```

### 💻 Quick Start

Gunakan template remote atau lokal dengan transisi yang mulus.

**1. Generate (*Scaffold*) dari remote repositori GitHub:**
```bash
symphony gen github.com/yourname/go-hexagonal-template@v1.2.0 --out ./my-api
```

**2. Preview *Dry-Run* sebelum menulis perubahan struktural ke *disk*:**
```bash
symphony gen github.com/yourname/go-blueprint --out ./my-api --dry-run
```

**3. Ulangi (*Reproduce*) environment secara mendetail dari arsip Lock File:**
```bash
# Jalankan langsung di dalam direktori project yang sebelumnya sudah di-generate
symphony re-gen 
```

**4. Bypass konfirmasi interaktif sepenuhnya (khusus integrasi CI/CD):**
```bash
symphony gen ./templates/base --out ./app --yes
```

### 🧠 Cara Kerja

Symphony berjalan di atas skema antarmuka (*Terminal UX*) yang sangat interaktif dalam metode kerja *Progressive 5-Fase* berikut:

1. **Discovery & Validation:** Melakukan validasi jalur direktori format (`template.yaml`) baik ke *path* GitHub, GitLab, atau arsip repositori komputer lokal; sekaligus mengecek kecocokan versi SDK dan kompilator perangkat Anda.
2. **Interactive Prompting:** Mengaktifkan sesi daftar-pertanyaan berantai untuk user secara estetik. (Contoh: Parameter isian '*Host URL Database*' hanya akan ditanyakan jika user mengaminkan opsi relasional-DB di tahapan sebelumnya).
3. **Pre-Execution Diff:** Menyelaraskan status-kode (*color-coded preview*) dan rekapitulasi lengkap terhadap modifikasi skema, melabeli status pada setiap file antara `[CREATE]`, `[MODIFY]`, hingga pengabaian `[SKIP]`.
4. **Generation & Injection:** Menyatukan rendering file statik standar (format `text/template`) berdampingan dengan pengeditan logika koding abstraktif (*ASTR/regex nodal mutations*), kemudian menuliskannya ke SSD secara paralel.
5. **Finalization:** Mengeksekusi modul-opsional skrip bash / shell *hook* pasca-scaffold (*mis: menjalankan `git init`, mengatur repositori `npm start`*) sebelum memancangkan konfigurasi `symphony.lock` ke dalam projek final tersebut.

### ⌨️ Referensi Command

| Command | Deskripsi Utama |
|---|---|
| `symphony gen <source>` | Melakukan inisiasi pembentukan struktur direktif program *from scratch* (lewat lokal *path*, URL arsip *template/Git*). |
| `symphony re-gen` | Pengulangan dan pencetakan replika *environment* persis merunut histori sistem pada berkas `symphony.lock`. |
| `symphony check <source>` | Menganalisa parameter kesehatan serta keselarasan logikal blueprint pada template (tanpa merepresentasikan hasil kode akhirnya). |
| `symphony list` | Merefleksikan entri *template* terbaru / tersimpan langsung dari arsip lokal komputer yang siap digunakan. |
| `symphony update <source>` | Memaksa penyegaran mutakhir / sinkronisasi terdepan *remote repositories template* sebelum dipakai. |
| `symphony cache clear` | Berfungsi menghapuskan *cache index* pada arsip Symphony internal mesin Anda. |

**Atribut (Global Flags):**
- `--out` / `-o`: Spesifikasi folder destinasi hasil *output*.
- `--dry-run` / `-d`: Operasi simulasi pratinjau yang diselenggarakan di *console terminal* secara tertutup.
- `--yes` / `-y`: Setuju ke dalam mode otomatis seutuhnya, tanpa pemberitahan iterasi lanjutan dari *Prompt*.
- `--config` / `-c`: Pengaturan konfigurasi opsi file khusus (`.symphony/config.yaml`).

### 📚 Dokumentasi

Jantung dari pergerakan instrumen Symphony berpusat pada blueprint `template.yaml`. Telusuri pemahaman modifikasi tingkat lanjut, hingga aturan main struktur *framework*:
- [Spesifikasi Sistem Template / Blueprint](docs/blueprint-spec.md)
- [Buku Panduan Kontributor (Penyumbang Kode)](docs/contributing.md)
- [Laporan Mutu (Reliability Release)](docs/QC_REPORT.md)
- [Detail Log Resolusi & Histori Update Sistem](CHANGELOG.md)

### 📄 Lisensi

Platform Symphony mematuhi kaidah sistem pendistribusian Perangkat Lunak Terbuka (Open Source) yang secara transparan tunduk di bawah deklarasi [LICENSE](LICENSE).