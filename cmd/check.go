package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Reivhell/symphony/internal/blueprint"
	"github.com/Reivhell/symphony/internal/remote"
	"github.com/Reivhell/symphony/internal/tui"
)

var checkJSON bool

type CheckReport struct {
	Source      string   `json:"source"`
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
	TemplateDir string   `json:"template_dir,omitempty"`
}

var checkCmd = &cobra.Command{
	Use:   "check <source>",
	Short: "Validasi status template secara komprehensif tanpa membuat skafolding",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		report := CheckReport{
			Source: source,
			Valid:  true,
		}

		if !checkJSON {
			fmt.Printf("🔍 Memeriksa template: %s\n", source)
		}

		addError := func(msg string) {
			report.Valid = false
			report.Errors = append(report.Errors, msg)
		}

		addWarn := func(msg string) {
			report.Warnings = append(report.Warnings, msg)
		}

		cacheDir, _ := remote.CacheDir()

		// 1 & 2. Remote Fetch & Basic Parse
		localPath, _, err := remote.Fetch(source, cacheDir)
		if err != nil {
			addError(fmt.Sprintf("Gagal mengakses source remote: %v", err))
			return renderReport(report)
		}
		
		report.TemplateDir = localPath

		bp, err := blueprint.Parse(localPath)
		if err != nil {
			addError(fmt.Sprintf("Parsing / Dependency Graph Error: %v", err))
			return renderReport(report)
		}

		// 3. Multilevel Inheritance Check (Circular Inheritance & Merge Prompts)
		wrapper := fetchWrapper{cacheDir: cacheDir}
		bp, err = blueprint.Resolve(bp, wrapper)
		if err != nil {
			addError(fmt.Sprintf("Kesalahan relasi/warisan template induk: %v", err))
		}

		// 4. Periksa kecocokan nama, versi
		if bp.SchemaVersion != "2" {
			addWarn(fmt.Sprintf("Schema_version yang diformat mungkin tidak dikenali: %s (sebaiknya gunakan '2')", bp.SchemaVersion))
		}

		// 5. Cek kelestarian file fisik referensi dari Action
		for i, action := range bp.Actions {
			if action.Type == "render" {
				if action.Source == "" {
					addError(fmt.Sprintf("Action Render ke-%d tidak memiliki property 'source'", i+1))
					continue
				}

				targetPths := filepath.Join(localPath, action.Source)
				if _, err := os.Stat(targetPths); os.IsNotExist(err) {
					addWarn(fmt.Sprintf("File source '%s' (dirujuk oleh action render) tidak ditemukan pada disk lokal", action.Source))
				}
			}
		}

		return renderReport(report)
	},
}

func renderReport(report CheckReport) error {
	if checkJSON {
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
		if !report.Valid {
			os.Exit(1)
		}
		return nil
	}

	fmt.Println()
	if report.Valid {
		fmt.Printf("  %s %s\n", tui.IconSuccess, tui.StyleSuccess.Render("Semua validasi (Parse & Dependency) sukses! Template siap digunakan."))
	} else {
		fmt.Printf("  %s %s\n", tui.IconError, tui.StyleDanger.Render("Terdapat Kesalahan Fatal Terdeteksi!"))
	}

	if len(report.Errors) > 0 {
		fmt.Println("\n  [Errors]:")
		for _, e := range report.Errors {
			fmt.Printf("    %s %s\n", tui.IconError, e)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\n  [Warnings]:")
		for _, w := range report.Warnings {
			fmt.Printf("    %s %s\n", tui.IconWarning, w)
		}
	}

	fmt.Println()
	if !report.Valid {
		os.Exit(1)
	}
	return nil
}

func init() {
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "Output pelaporan dalam format JSON (machine-readable)")
	rootCmd.AddCommand(checkCmd)
}
