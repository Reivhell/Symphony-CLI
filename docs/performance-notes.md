# Performance notes (Phase 05)

## CLI startup (`symphony --help`)

Measure locally (bash):

```bash
for i in $(seq 1 10); do time ./bin/symphony --help >/dev/null 2>&1; done
```

On Windows PowerShell, use:

```powershell
1..10 | ForEach-Object { Measure-Command { .\bin\symphony.exe --help | Out-Null } }
```

Target: interactive perception under ~100ms cold start is ideal; actual numbers depend on disk and AV.

## Renderer benchmarks

```bash
go test ./internal/engine/... -bench=. -benchmem -count=3
```

Record the latest figures in `docs/QC_REPORT.md` when cutting a release.
