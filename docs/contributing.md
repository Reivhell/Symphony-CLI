# Contributing to Symphony

Thank you for improving Symphony. This guide covers local setup, tests, commit conventions, and pull requests.

---

## Development setup

1. Install **Go 1.24+** (see `go.mod` for the toolchain used in CI).
2. Clone the repository.
3. Install tools (optional but recommended):
   - [golangci-lint](https://golangci-lint.run/welcome/install/)
   - [GoReleaser](https://goreleaser.com/install/) (for release snapshots)
4. From the repo root:

   ```bash
   go mod download
   make build
   make test
   ```

---

## Makefile targets

| Target | Purpose |
|--------|---------|
| `make all` | `lint`, `test`, `build` |
| `make build` | Produce `./bin/symphony` with version ldflags |
| `make test` | Run tests with `-race` |
| `make test-cover` | Coverage report (`coverage.html`) |
| `make lint` | `golangci-lint run ./...` |
| `make tidy` | `go mod tidy` |
| `make clean` | Remove `bin/`, coverage artifacts, `dist/` |
| `make snapshot` | `goreleaser release --snapshot --clean` |
| `make install` | `go install` with ldflags |

---

## Tests

```bash
go test ./... -race
```

Add tests for new packages and for behavior changes in `internal/engine`, `internal/blueprint`, `internal/tui`, etc.

---

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation only
- `chore:` tooling, CI, no production code change
- `test:` tests only
- `perf:` performance

Example: `feat: add plugin glob matching for render actions`

---

## Pull requests

1. Open a PR against the default branch (e.g. `main`).
2. Describe **what** changed and **why** (motivation / context).
3. Ensure `make test` and `make lint` pass locally.
4. Keep changes focused; unrelated refactors belong in separate PRs.
5. Respond to review feedback; maintainers may request tests or doc updates.

---

## Release process (maintainers)

1. Tag with `v*` (e.g. `v1.2.3`).
2. CI (`release` workflow) runs GoReleaser to publish GitHub release artifacts.
3. Replace placeholder `username/symphony` in `README.md`, `.goreleaser.yaml`, and `install.sh` with the real org/repo before publishing.

---

## Code style

- Run `gofmt` / `goimports` (enforced via `golangci-lint`).
- Prefer small, readable functions; handle errors explicitly (no silent failures).
- Do not commit secrets or tokens.
