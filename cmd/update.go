// Package cmd — update.go
// Perintah: symphony update <source>
// Fungsi: Memperbarui template yang ter-cache ke versi terbaru dari remote.
//
// TODO: Implementasi di Phase 3 (remote fetching & caching)
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <source>",
	Short: "Update template ke versi terbaru",
	Long: `Memperbarui template yang tersimpan di cache lokal ke versi terbaru
yang tersedia di remote source-nya.

Contoh:
  symphony update github.com/user/go-blueprint`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[update] belum diimplementasikan — Coming in Phase 3")
		fmt.Printf("  Source: %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
