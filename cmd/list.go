package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/Reivhell/symphony/internal/config"
	"github.com/Reivhell/symphony/internal/remote"
	"github.com/Reivhell/symphony/internal/tui"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Menampilkan template lokal dan riwayat templat di cache (tersimpan)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil && outputFormat == "human" {
			// Print warning saja, biarkan lanjut pake Default config.
			fmt.Printf("  %s %s\n\n", tui.IconWarning, tui.StyleWarning.Render(err.Error()))
		}

		type templateInfo struct {
			Type     string `json:"type"`     // "local" atau "cache"
			Source   string `json:"source"`
			Size     string `json:"size"`
			Age      string `json:"age"`      
			IsTagged bool   `json:"is_tagged"`
		}

		var templates []templateInfo

		// 1. Muat Local Config Template
		for _, lp := range cfg.LocalTemplatePaths {
			sizeBytes, _ := dirSize(lp)
			templates = append(templates, templateInfo{
				Type:   "local",
				Source: lp,
				Size:   formatBytes(sizeBytes),
				Age:    "Penyimpanan Local",
			})
		}

		// 2. Muat Cache System
		cacheDir, _ := remote.CacheDir()
		if cacheDir != "" {
			entries, err := remote.List(cacheDir)
			if err == nil {
				for _, e := range entries {
					templates = append(templates, templateInfo{
						Type:     "cache",
						Source:   e.Source,
						Size:     formatBytes(e.SizeBytes),
						Age:      formatTimeAgo(e.CachedAt),
						IsTagged: e.IsTagged,
					})
				}
			}
		}

		// 3. Tampilkan Laporan Output
		if checkJSON || listJSON || outputFormat == "json" {
			b, _ := json.MarshalIndent(templates, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Printf("\n📦 %s\n", tui.StyleHeader.Render("Symphony Templates (Ditemukan: "+fmt.Sprint(len(templates))+")"))
		fmt.Println()

		if len(templates) == 0 {
			fmt.Println("  Tidak ada templat lokal maupun cache yang tersedia.")
			return nil
		}

		fmt.Printf(
			"  %-10s | %-45s | %-8s | %s\n",
			tui.StyleMuted.Render("TYPE"),
			tui.StyleMuted.Render("SOURCE / PATH"),
			tui.StyleMuted.Render("SIZE"),
			tui.StyleMuted.Render("CACHED / STATUS"),
		)
		fmt.Println("  " + tui.StyleDivider.Render("---------------------------------------------------------------------------------------"))

		for _, t := range templates {
			typStr := t.Type
			if typStr == "local" {
				typStr = tui.StyleSuccess.Render(typStr)
			} else {
				typStr = tui.StyleBrand.Render(typStr)
			}

			status := t.Age
			if t.IsTagged {
				status += " (Tagged)"
			}

			// Format padding manual
			fmt.Printf("  %-19s| %-45s | %-8s | %s\n", typStr, shortenStr(t.Source, 45), t.Size, status)
		}
		fmt.Println()
		
		return nil
	},
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff.Hours() > 24:
		return fmt.Sprintf("%d hr lalu", int(diff.Hours()/24))
	case diff.Hours() >= 1:
		return fmt.Sprintf("%d jam lalu", int(diff.Hours()))
	case diff.Minutes() >= 1:
		return fmt.Sprintf("%d mnt lalu", int(diff.Minutes()))
	default:
		return "baru saja"
	}
}

func shortenStr(s string, limit int) string {
	if len(s) > limit {
		return s[:limit-3] + "..."
	}
	return s
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Keluarkan data berformat JSON Array")
	rootCmd.AddCommand(listCmd)
}
