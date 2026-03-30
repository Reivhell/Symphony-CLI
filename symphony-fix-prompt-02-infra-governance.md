# Symphony CLI — Fix Prompt 02: Repository, Infrastructure & Governance
> **Tipe Dokumen:** AI Execution Prompt
> **Sesi:** 2 dari 2
> **Cakupan:** Repository hygiene, .gitignore, distribution config, CI/CD pipeline, governance files, hardening
> **Referensi:** `symphony-remediation-blueprint.md` — Issue #4, #5, #6, #7, #8, #10, #12, #13, #14, Improvisasi 16.2, 16.4
> **Commit Target:** `chore/fix-02-repository-infra-governance`
> **Prasyarat:** Sesi 1 (`symphony-fix-prompt-01-code-fixes.md`) sudah selesai dan semua test hijau

---

## Konteks untuk AI

Kamu adalah senior DevOps engineer yang melanjutkan pekerjaan dari Sesi 1. Build sudah bersih, test sudah hijau, dan module path sudah benar. Tugas kamu di sesi ini adalah membersihkan repositori dari artefak yang tidak seharusnya ada, memperbaiki seluruh konfigurasi distribusi yang masih menggunakan placeholder, membangun pipeline CI yang benar-benar berjalan, dan menambahkan file governance yang diperlukan agar repositori ini layak dilihat publik.

Aturan yang wajib dipatuhi selama sesi ini:

Pertama, kerjakan setiap tugas **secara berurutan**. Beberapa tugas (khususnya Tugas 2 dan 3) harus diselesaikan sebelum tugas berikutnya bisa diverifikasi. Kedua, semua file yang dibuat atau dimodifikasi harus menggunakan path yang benar `github.com/Reivhell/symphony` — bukan `github.com/username/symphony`. Ketiga, jangan menyentuh kode Go (`*.go`) di sesi ini kecuali ada instruksi eksplisit. Fokus sepenuhnya pada konfigurasi, infrastruktur, dan file teks.

---

## Tugas 1 — Perbaiki `.gitignore` Secara Menyeluruh

**Masalah:** `.gitignore` saat ini tidak mencakup pola-pola yang diperlukan untuk proyek Go, terbukti dari keberadaan `symphony.exe`, `bin/`, `debug.log`, `err.log`, dan `expr_test.log` di repositori.

**Yang harus dilakukan:**

Ganti seluruh konten `.gitignore` di root repositori dengan konfigurasi komprehensif berikut:

```gitignore
# ─── Binary & Build Output ─────────────────────────────────────
/bin/
/dist/
*.exe
*.exe~
*.dll
*.so
*.dylib

# ─── Test & Coverage Artifacts ─────────────────────────────────
*.test
*.out
coverage.out
coverage.html
coverage.txt

# ─── Log Files ─────────────────────────────────────────────────
*.log

# ─── Go Workspace (local override only) ────────────────────────
go.work
go.work.sum

# ─── Environment & Secrets ─────────────────────────────────────
.env
.env.local
.env.*.local
*.pem
*.key
*.cert

# ─── OS Artifacts ───────────────────────────────────────────────
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db
desktop.ini

# ─── Editor & IDE ───────────────────────────────────────────────
.idea/
.vscode/
*.swp
*.swo
*~
.vim/
*.code-workspace

# ─── Temporary Files ────────────────────────────────────────────
tmp/
temp/
*.tmp

# ─── Symphony-specific ──────────────────────────────────────────
# Hasil scaffold dari testing lokal
/testdata/output/
# Debug scripts sementara
bootstrap.ps1
```

**Verifikasi setelah selesai:**
```bash
git status
# Pastikan .gitignore muncul sebagai modified
# Pastikan tidak ada file baru yang muncul sebagai untracked yang seharusnya diabaikan
```

---

## Tugas 2 — Hapus Artefak Binary dan Log dari Version Control

**Masalah:** `symphony.exe`, direktori `bin/`, `debug.log`, `err.log`, dan `expr_test.log` sudah terlanjur di-track oleh Git. Memperbaiki `.gitignore` saja tidak cukup — file-file yang sudah di-track harus secara eksplisit di-untrack menggunakan `git rm --cached`.

**Yang harus dilakukan:**

Langkah 2a — Hapus file-file tersebut dari Git tracking (bukan dari disk):

```bash
# Hapus binary Windows
git rm --cached symphony.exe

# Hapus direktori bin (seluruh isinya)
git rm --cached -r bin/

# Hapus semua log file yang ter-track
git rm --cached debug.log err.log expr_test.log

# Jika ada file .log lain yang ter-track, hapus sekaligus:
git ls-files "*.log" | xargs git rm --cached
```

Langkah 2b — Verifikasi bahwa file-file tersebut sudah masuk ke staging area sebagai deletion:

```bash
git status
# Expected: file-file di atas muncul sebagai "deleted" di staging area
# File fisiknya masih ada di disk (yang kita hapus hanya tracking-nya)
```

Langkah 2c — Evaluasi `bootstrap.ps1`. Buka file tersebut dan tentukan tujuannya. Jika ini adalah script setup development Windows yang masih digunakan, pindahkan ke folder `scripts/` dan perbarui referensinya di dokumentasi. Jika ini adalah artifact debugging yang sudah tidak relevan, tambahkan ke `.gitignore` dan hapus dari tracking. Jika ragu, tambahkan ke `.gitignore` terlebih dahulu:

```bash
# Jika akan dihapus dari tracking:
git rm --cached bootstrap.ps1
```

---

## Tugas 3 — Perbaiki `install.sh` agar Tidak Broken by Default

**Masalah:** `install.sh` baris 10 menggunakan fallback ke `username/symphony` yang tidak ada:

```sh
REPO="${SYMPHONY_GITHUB_REPO:-username/symphony}"
```

Ini berarti setiap orang yang menjalankan install script tanpa environment variable khusus akan mendapat error karena mencoba mengunduh dari repositori yang tidak ada.

**Yang harus dilakukan:**

Langkah 3a — Buka `install.sh` dan cari baris yang mendefinisikan `REPO`. Ganti dengan path repositori yang benar:

```sh
# Sebelum:
REPO="${SYMPHONY_GITHUB_REPO:-username/symphony}"

# Sesudah:
REPO="${SYMPHONY_GITHUB_REPO:-Reivhell/Symphony-CLI}"
```

Langkah 3b — Periksa seluruh isi `install.sh` untuk kemunculan lain dari `username/symphony` atau `github.com/username`. Ganti semua kemunculan tersebut dengan path yang benar:

```bash
grep -n "username" install.sh
# Setiap hasil harus diperbaiki
```

Langkah 3c — Verifikasi bahwa `install.sh` valid secara sintaksis:

```bash
sh -n install.sh
# Expected: tidak ada output (silent success = valid syntax)
```

---

## Tugas 4 — Perbaiki `.goreleaser.yaml` dari Placeholder ke Identitas Nyata

**Masalah:** `.goreleaser.yaml` masih menggunakan `username/symphony` di beberapa tempat kritis: `ldflags`, `release.github`, dan `brews`.

**Yang harus dilakukan:**

Buka `.goreleaser.yaml` dan lakukan perubahan berikut secara teliti. Setiap baris yang diubah ditandai dengan komentar `# CHANGED`:

```yaml
# .goreleaser.yaml — versi yang diperbaiki (hanya tampilkan bagian yang berubah)

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
      - -s
      - -w
      # CHANGED: dari github.com/username/symphony ke github.com/Reivhell/symphony
      - -X github.com/Reivhell/symphony/cmd.Version={{.Version}}
      - -X github.com/Reivhell/symphony/cmd.Commit={{.Commit}}
      - -X github.com/Reivhell/symphony/cmd.BuildDate={{.Date}}

release:
  github:
    # CHANGED: dari username/symphony ke Reivhell/Symphony-CLI
    owner: Reivhell
    name: Symphony-CLI
  draft: false
  prerelease: auto
  name_template: "Symphony v{{.Version}}"
  footer: |
    ## Installation
    ```bash
    # CHANGED: URL yang benar
    curl -sSL https://raw.githubusercontent.com/Reivhell/Symphony-CLI/main/install.sh | sh
    ```

brews:
  - name: symphony
    # CHANGED: URL yang benar
    homepage: "https://github.com/Reivhell/Symphony-CLI"
    description: "The Adaptive Scaffolding Engine"
    skip_upload: true
    repository:
      # CHANGED: owner dan name yang benar
      owner: Reivhell
      name: homebrew-symphony
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
    install: |
      bin.install "symphony"
    test: |
      system "#{bin}/symphony", "version"
```

**Verifikasi setelah selesai:**
```bash
# Pastikan tidak ada 'username' yang tersisa
grep -n "username" .goreleaser.yaml
# Expected: tidak ada output

# Validasi syntax goreleaser (jika goreleaser terinstal)
goreleaser check
```

---

## Tugas 5 — Perbarui `README.md` dengan Informasi yang Akurat

**Yang harus dilakukan:**

Langkah 5a — Buka `README.md` dan perbaiki semua URL instalasi. Ganti semua kemunculan `username/symphony` dan URL release lama:

```markdown
## Installation

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/Reivhell/Symphony-CLI/main/install.sh | sh
```

**Windows:** Download `symphony_*.zip` dari [GitHub Releases](https://github.com/Reivhell/Symphony-CLI/releases), extract `symphony.exe`, dan tambahkan ke `PATH`.

**From source:**
```bash
go install github.com/Reivhell/symphony@latest
```

Langkah 5b — Tambahkan status badges di bagian atas README, tepat di bawah judul utama dan deskripsi singkat:

```markdown
[![CI](https://github.com/Reivhell/Symphony-CLI/actions/workflows/ci.yml/badge.svg)](https://github.com/Reivhell/Symphony-CLI/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Reivhell/symphony)](https://goreportcard.com/report/github.com/Reivhell/symphony)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

Catatan: badge CI akan menampilkan status "failing" atau "unknown" sampai GitHub Actions workflow dibuat di Tugas 6. Ini normal — badge akan otomatis menjadi hijau setelah CI berjalan pertama kali.

---

## Tugas 6 — Buat GitHub Actions CI Pipeline

**Masalah:** Pipeline CI yang seharusnya mencegah kode bermasalah masuk ke `main` tidak berjalan, terbukti dari build failure dan failing test yang berhasil masuk ke branch utama.

**Yang harus dilakukan:**

Periksa apakah `.github/workflows/ci.yml` sudah ada. Jika sudah ada, evaluasi isinya — jika tidak mencakup pemeriksaan yang diperlukan, ganti seluruhnya. Jika belum ada, buat file baru. Konten yang benar adalah sebagai berikut:

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: ["main", "develop", "feature/**", "fix/**"]
  pull_request:
    branches: ["main"]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  build-and-test:
    name: Build & Test (${{ matrix.os }} / Go ${{ matrix.go }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ["1.22", "1.23"]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go ${{ matrix.go }}
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Verify go.mod is consistent
        run: go mod verify

      - name: Run go vet
        run: go vet ./...

      - name: Build binary
        run: go build ./...

      - name: Run tests with race detector
        run: go test ./... -race -count=1 -timeout=120s

      - name: Validate no placeholder strings (Linux only)
        if: runner.os == 'Linux'
        run: |
          if grep -rn "username/symphony" \
               --include="*.go" \
               --include="*.yaml" \
               --include="*.sh" \
               --exclude-dir=".git" . 2>/dev/null; then
            echo "FAIL: Placeholder module path still exists in codebase"
            exit 1
          fi
          echo "OK: No placeholder strings found"

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  coverage:
    name: Test Coverage
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true
      - name: Generate coverage report
        run: |
          go test ./... -coverprofile=coverage.out -covermode=atomic
          go tool cover -func=coverage.out | tail -1
      - name: Upload coverage to Codecov (optional)
        uses: codecov/codecov-action@v4
        continue-on-error: true
        with:
          file: coverage.out
          fail_ci_if_error: false
```

Juga buat file workflow terpisah untuk release:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  goreleaser:
    name: GoReleaser
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Verifikasi setelah selesai:**

Verifikasi file valid secara YAML:
```bash
# Jika yq atau python tersedia:
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "Valid YAML"
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "Valid YAML"
```

---

## Tugas 7 — Buat `SECURITY.md`

**Yang harus dilakukan:**

Buat file `SECURITY.md` di root repositori dengan konten berikut:

```markdown
# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ Yes    |

## Reporting a Vulnerability

Symphony CLI processes templates from external sources and can execute shell
commands via post-scaffold hooks. Security is a top priority.

**Do NOT create a public GitHub Issue for security vulnerabilities.**

Please report security vulnerabilities privately using **GitHub Security Advisories**:

1. Go to https://github.com/Reivhell/Symphony-CLI/security/advisories
2. Click "New draft security advisory"
3. Describe the vulnerability, steps to reproduce, and potential impact

We commit to responding within **48 hours** and releasing a patch within
**7 days** for confirmed vulnerabilities.

## Security Scope

**In-scope (please report):**
- Expression evaluator injection (`pkg/expr`) — template `if:` conditions
- Path traversal in file writer (`internal/engine/writer.go`)
- Malicious post-scaffold hook execution
- Remote template fetching vulnerabilities
- Arbitrary code execution via template rendering

**Out-of-scope:**
- Security of third-party templates (responsibility of template authors)
- Vulnerabilities in dependencies (report to respective maintainers)

## Security Hardening Notes for Template Authors

Symphony executes shell commands defined in template `hooks`. Users are
encouraged to always review `template.yaml` — especially the `actions` section
with `type: shell` — before running `symphony gen` on untrusted templates.
```

---

## Tugas 8 — Buat GitHub Issue & PR Templates

**Yang harus dilakukan:**

Buat direktori `.github/ISSUE_TEMPLATE/` jika belum ada. Kemudian buat tiga file berikut:

**File 1: `.github/ISSUE_TEMPLATE/bug_report.md`**

```markdown
---
name: Bug Report
about: Laporkan bug atau perilaku yang tidak sesuai ekspektasi
title: "[BUG] "
labels: ["bug", "needs-triage"]
assignees: ""
---

## Deskripsi Bug

Jelaskan bug secara singkat dan jelas.

## Langkah Reproduksi

1. Jalankan perintah: `symphony ...`
2. Gunakan template: `...`
3. Lihat error

## Perilaku yang Diharapkan

Jelaskan apa yang seharusnya terjadi.

## Output Aktual

```
Tempelkan output error atau pesan yang salah di sini
```

## Environment

- **Symphony version:** (output dari `symphony version`)
- **OS:** (Linux / macOS / Windows)
- **Architecture:** (amd64 / arm64)
- **Go version (jika build dari source):** 

## Informasi Tambahan

Tambahkan informasi lain yang relevan, screenshot, atau konten `template.yaml` yang digunakan.
```

**File 2: `.github/ISSUE_TEMPLATE/feature_request.md`**

```markdown
---
name: Feature Request
about: Usulkan fitur baru atau peningkatan yang ada
title: "[FEAT] "
labels: ["enhancement"]
assignees: ""
---

## Ringkasan Fitur

Jelaskan fitur yang kamu usulkan dalam satu paragraf.

## Motivasi & Use Case

Mengapa fitur ini diperlukan? Masalah apa yang diselesaikan?

## Deskripsi Solusi yang Diinginkan

Jelaskan solusi yang kamu bayangkan. Jika relevan, sertakan contoh konfigurasi `template.yaml` atau perintah CLI.

## Alternatif yang Sudah Dipertimbangkan

Apakah ada cara lain untuk mencapai hasil yang sama dengan tools yang sudah ada?

## Informasi Tambahan

Tambahkan konteks, referensi, atau contoh dari tools serupa jika ada.
```

**File 3: `.github/PULL_REQUEST_TEMPLATE.md`**

```markdown
## Ringkasan

Jelaskan perubahan yang dilakukan dalam PR ini dan mengapa perubahan ini diperlukan.

Closes # (nomor issue yang diselesaikan, jika ada)

## Jenis Perubahan

- [ ] Bug fix (perubahan non-breaking yang memperbaiki issue)
- [ ] New feature (perubahan non-breaking yang menambahkan fungsionalitas)
- [ ] Breaking change (fix atau feature yang menyebabkan fungsionalitas yang ada tidak bekerja seperti sebelumnya)
- [ ] Documentation update
- [ ] Refactoring / code cleanup

## Checklist Sebelum Merge

- [ ] `go test ./... -race` lulus di mesin lokal
- [ ] `go vet ./...` tidak menghasilkan output (clean)
- [ ] Test baru sudah ditambahkan untuk perubahan yang dibuat
- [ ] Tidak ada string placeholder `username/symphony` yang tersisa
- [ ] Dokumentasi (`docs/`, `README.md`) sudah diperbarui jika diperlukan
- [ ] Commit messages mengikuti format Conventional Commits (`feat:`, `fix:`, `chore:`, dst)
```

---

## Tugas 9 — Tambahkan `.github/CODEOWNERS`

**Yang harus dilakukan:**

Buat file `.github/CODEOWNERS` untuk memastikan perubahan pada file-file kritis memerlukan review dari maintainer utama:

```
# CODEOWNERS — Tentukan siapa yang harus me-review perubahan pada file kritis

# Default: semua file memerlukan review
*                          @Reivhell

# File kritis yang selalu memerlukan review eksplisit
go.mod                     @Reivhell
go.sum                     @Reivhell
.goreleaser.yaml           @Reivhell
.github/workflows/         @Reivhell
internal/engine/           @Reivhell
pkg/expr/                  @Reivhell
```

---

## Tugas 10 — Commit Semua Perubahan Sesi Ini

Setelah semua tugas di atas selesai, lakukan satu verifikasi menyeluruh sebelum commit:

```bash
# 1. Pastikan tidak ada artefak yang masih ter-track
git ls-files "*.log" "*.exe"
# Expected: tidak ada output

# 2. Pastikan tidak ada placeholder yang tersisa
grep -r "username/symphony" . --include="*.go" --include="*.yaml" --include="*.sh" --exclude-dir=".git"
# Expected: tidak ada output

# 3. Pastikan build masih bersih setelah semua perubahan
go build ./...
go test ./... -race

# 4. Lihat semua perubahan yang akan di-commit
git status
git diff --stat
```

Kemudian commit dengan pesan yang deskriptif:

```bash
git add -A
git commit -m "chore: repository cleanup, CI pipeline, and governance files

- chore: update .gitignore with comprehensive Go project patterns
- chore: untrack binary artifacts (symphony.exe, bin/) and log files
- fix: update install.sh default REPO from username/symphony to Reivhell/Symphony-CLI
- fix: update .goreleaser.yaml owner, name, and all URLs from placeholder to real identity
- docs: update README.md installation URLs and add CI/license badges
- ci: add GitHub Actions CI pipeline (build-and-test matrix + lint + coverage)
- ci: add GitHub Actions release pipeline (goreleaser on tag push)
- docs: add SECURITY.md with vulnerability reporting guidelines
- chore: add GitHub issue templates (bug report, feature request)
- chore: add GitHub PR template with merge checklist
- chore: add CODEOWNERS for critical files"
```

---

## Tugas 11 — Instruksi Manual: Branch Protection di GitHub

Langkah ini tidak bisa dilakukan melalui kode — harus dikonfigurasi langsung di GitHub web interface setelah semua commit di-push.

Setelah semua commit dari Sesi 1 dan Sesi 2 di-push ke `main`, lakukan langkah berikut di GitHub:

Pertama, buka halaman repositori di `https://github.com/Reivhell/Symphony-CLI`. Kedua, masuk ke **Settings** → **Branches** → klik **Add branch protection rule**. Ketiga, isi **Branch name pattern** dengan `main`. Keempat, aktifkan opsi-opsi berikut:

**Require a pull request before merging** — aktifkan untuk mencegah push langsung ke main di masa depan.

**Require status checks to pass before merging** — aktifkan, kemudian tambahkan `build-and-test (ubuntu-latest, 1.22)` sebagai required status check (job ini akan muncul setelah CI pertama kali berjalan).

**Do not allow bypassing the above settings** — aktifkan agar aturan berlaku bahkan untuk owner repositori.

Kelima, klik **Create**.

---

## Tugas 12 — Instruksi Manual: Daftarkan ke Go Report Card

Langkah ini dilakukan secara manual setelah branch protection aktif dan CI sudah berjalan hijau:

Buka browser dan kunjungi `https://goreportcard.com`. Di kotak pencarian, masukkan `github.com/Reivhell/symphony` dan klik **Check**. Go Report Card akan menganalisis kodebase dan menghasilkan badge. Salin URL badge yang dihasilkan dan tambahkan ke README jika belum ada.

---

## Checklist Penyelesaian Sesi 2

Verifikasi kondisi berikut sebelum menutup sesi ini:

**Repository Hygiene**
- [ ] `git ls-files "*.log" "*.exe"` — tidak ada output
- [ ] `git ls-files bin/` — tidak ada output
- [ ] `.gitignore` mencakup semua pola binary dan log

**Configuration**
- [ ] `install.sh` menggunakan `Reivhell/Symphony-CLI` sebagai default REPO
- [ ] `.goreleaser.yaml` tidak mengandung `username` di manapun
- [ ] `README.md` menampilkan URL instalasi yang akurat

**CI/CD**
- [ ] `.github/workflows/ci.yml` ada dan valid syntax-nya
- [ ] `.github/workflows/release.yml` ada dan valid syntax-nya
- [ ] Setelah push, CI berjalan dan semua job hijau

**Governance**
- [ ] `SECURITY.md` ada di root repositori
- [ ] `.github/ISSUE_TEMPLATE/bug_report.md` ada
- [ ] `.github/ISSUE_TEMPLATE/feature_request.md` ada
- [ ] `.github/PULL_REQUEST_TEMPLATE.md` ada
- [ ] `.github/CODEOWNERS` ada

**Manual Steps (dikonfirmasi setelah push)**
- [ ] Branch protection aktif di `main` dengan required status checks
- [ ] Repositori terdaftar di Go Report Card

---

## Verifikasi Akhir — Kedua Sesi Selesai

Setelah Sesi 1 dan Sesi 2 keduanya selesai, jalankan verifikasi akhir menyeluruh:

```bash
# Clean build tanpa warning
go build ./...

# Semua test lulus dengan race detector
go test ./... -race -count=1

# Tidak ada placeholder yang tersisa
grep -r "username/symphony" . --include="*.go" --include="*.yaml" --include="*.sh" --exclude-dir=".git"

# Tidak ada artefak di version control
git ls-files "*.log" "*.exe" bin/

# Makefile check berjalan bersih
make check
make check-placeholders
```

Jika semua perintah di atas menghasilkan output yang benar (build sukses, test lulus, grep tidak menemukan hasil, git ls-files tidak menemukan artefak), maka repositori Symphony CLI telah berhasil diperbaiki dan siap untuk menerima feature development dari `symphony-creative-update-blueprint.md`.

---

## Catatan Penting untuk AI

Sesi ini bersifat **infrastruktur dan konfigurasi** — jangan mengubah logika kode Go yang sudah diperbaiki di Sesi 1. Jika selama proses ini kamu menemukan file Go yang masih perlu perubahan (misalnya komentar yang masih menyebut `username/symphony`), lakukan perubahan tersebut tetapi pisahkan dalam commit terpisah dengan label yang jelas agar tidak mencampur scope antara code fix dan infra fix. Prioritas utama adalah memastikan bahwa setelah kedua sesi ini selesai, repositori berada dalam kondisi yang benar-benar bersih dan bisa dilihat oleh siapapun tanpa rasa malu.

Use Context7  For Newest Knowledge