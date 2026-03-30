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
