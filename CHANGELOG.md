# Changelog

All notable changes to Symphony will be documented in this file.

## v0.1.0 — 2026-03-29

### Added

- Phase 05 quality control: path traversal protection for writes, expression length limits, template value hardening for user prompts, plugin process handling, write session rollback on cancel, file locking for template cache, improved network error messages.
- `docs/performance-notes.md`, `docs/E2E_SCENARIOS.md`, `docs/QC_REPORT.md` (release gate artifacts).
- Benchmarks for renderer (`internal/engine/renderer_bench_test.go`).
- Expanded `testdata/templates/hexagonal-go` fixture for E2E-style checks.
- `testdata/templates/broken-yaml` invalid fixture for `symphony check` testing.

### Changed

- `engine.Execute` now accepts `context.Context` for cancellation; hooks use `exec.CommandContext`.
- CLI `gen` / `re-gen` exit with code 4 on context cancellation after rollback attempt.

### Security

- Reject write targets outside `OutputDir` (`ErrPathTraversal`).
- Escape delimiter sequences in user string values passed to `text/template` data (defense in depth).
