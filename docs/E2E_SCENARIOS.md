# End-to-end scenario checklist (Phase 05)

Run these manually (or automate in CI where practical) before release. Update the table when you verify.

| Scenario | Status | Exit Code | Notes |
|----------|--------|-----------|-------|
| 1: Happy path — `symphony gen ./testdata/templates/hexagonal-go --out /tmp/e2e-1 --yes` | Pending | | Expect `symphony.lock`, hooks, 0 |
| 2: Dry-run — same with `--dry-run` | Pending | | No files under output |
| 3: Conditional — template with `if`; answer so branch skipped | Pending | | Skipped files absent |
| 4: Re-gen — gen then delete file then `symphony re-gen --yes` | Pending | | Restored files |
| 5: Invalid input — regex validation failure | Pending | | Clear error, no panic |
| 6: Plugin template | Pending | | Plugin invoked |
| 7a: `symphony check` valid template | Pending | | Pass |
| 7b: `symphony check ./testdata/templates/broken-yaml` | Pending | | Fails with descriptive error |

Commands use POSIX paths; on Windows adjust `/tmp` to a temp folder or use `%TEMP%`.
