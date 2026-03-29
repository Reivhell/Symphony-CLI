// Symphony CLI — The Adaptive Scaffolding Engine
// Entry point utama. Semua logika dimulai dari cmd/root.go
package main

import (
	"github.com/username/symphony/cmd"
	"github.com/username/symphony/internal/version"
)

func main() {
	version.Version = cmd.Version
	version.Commit = cmd.Commit
	version.BuildDate = cmd.BuildDate
	cmd.Execute()
}
