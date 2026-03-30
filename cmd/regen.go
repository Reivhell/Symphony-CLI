package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Reivhell/symphony/internal/blueprint"
	"github.com/Reivhell/symphony/internal/engine"
	"github.com/Reivhell/symphony/internal/lock"
	"github.com/Reivhell/symphony/internal/remote"
	"github.com/Reivhell/symphony/internal/tui"
)

var (
	regenDryRun  bool
	regenYes     bool
	regenNoHooks bool
)

var regenCmd = &cobra.Command{
	Use:   "re-gen",
	Short: "Ulangi skafolding secara identik berdasarkan symphony.lock",
	Long: `Membaca symphony.lock di direktori saat ini dan mengulang proses
scaffolding secara deterministik — dengan konfigurasi, jawaban input, dan template
yang sama persis atau yang telah diperbarui bila templat upstream berubah.

Berguna ketika Anda tidak sengaja menghapus folder tertentu, atau 
ingin menjalankan injeksi syntax baru yang ada di update terbaru versi template aslinya.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// 1. Baca Lockfile
		lockFile, err := lock.Read(cwd)
		if err != nil {
			return fmt.Errorf("symphony.lock tidak ditemukan (%w). Pastikan dijalankan di dalam root proyek ter-scaffold", err)
		}

		// 2. Tentukan Origin Source Template (& Versi rilisnya bila ada)
		source := lockFile.Template.Source
		// Jika dia remote dengan tagged versioning
		if !checkLocal(source) && lockFile.Template.Version != "" {
			source = source + "@" + lockFile.Template.Version
		}

		// 3. Cache & Fetch (Unduh jika versi / source belum di pc kita)
		cacheDir, err := remote.CacheDir()
		if err != nil {
			return err
		}

		absSource, fetchMeta, err := remote.Fetch(source, cacheDir)
		if err != nil {
			return fmt.Errorf("gagal mengunduh atau mereferensikan sumber templat %s: %w", source, err)
		}

		// 4. Parser & Resolver
		bp, err := blueprint.Parse(absSource)
		if err != nil {
			return err
		}

		wrapper := fetchWrapper{cacheDir: cacheDir}
		bp, err = blueprint.Resolve(bp, wrapper)
		if err != nil {
			return fmt.Errorf("gagal menyelesaikan pewarisan templat yang dirujuk ulang: %w", err)
		}
		if fetchMeta.Commit != "" {
			bp.Version = fetchMeta.Version + "+" + fetchMeta.Commit
		}

		if outputFormat == "human" {
			vString := lockFile.Template.Version
			if vString == "" {
				vString = "latest"
			}
			fmt.Printf("🚀 Melakukan Re-Scaffolding %s (template revisi: %s)\n\n", tui.StyleHighlight.Render(bp.Name), vString)
		}

		// 5. Build Engine Context menggunakan Input tersimpan
		var ws *engine.WriteSession
		if !regenDryRun {
			ws = engine.NewWriteSession()
		}
		ctx := &engine.EngineContext{
			Values:       lockFile.Inputs,
			SourceDir:    absSource,
			OutputDir:    cwd,
			DryRun:       regenDryRun,
			NoHooks:      regenNoHooks,
			YesAll:       regenYes, // Auto set ke state konfirmasi dari parameter run
			Plugins:      engine.PluginRenderersFromBlueprint(bp),
			ExecCtx:      cmd.Context(),
			WriteSession: ws,
			Format:       outputFormat,
		}

		// 6. Validasi 
		if err := engine.ValidateInputs(bp, ctx); err != nil {
			return fmt.Errorf("konfigurasi lockfile tidak lolos validasi input template ter-update: %w", err)
		}

		walkerEng := engine.New(bp, ctx)
		actions, err := walkerEng.Prepare()
		if err != nil {
			return err
		}

		totalFiles := 0
		for _, a := range actions {
			if a.ShouldRun && a.Original.Type != "shell" {
				totalFiles++
			}
		}

		if outputFormat == "human" {
			ctx.Reporter = tui.NewTUIReporter(totalFiles)
		} else if outputFormat == "json" {
			ctx.Reporter = tui.NewJSONReporter(totalFiles)
		}

		// TUI Preview Jika output command bukan JSON atau Yes All diset Off
		if outputFormat == "human" && !regenYes {
			confirmed, err := tui.RenderSummary(actions)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println(tui.StyleMuted.Render("\n  Operasi Re-Generate Dibatalkan."))
				os.Exit(4)
			}
		}

		// 7. Re-Execute Scaffolding
		if err := walkerEng.Execute(cmd.Context(), actions); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Fprintln(os.Stderr, "Operation cancelled; rolled back partial output where possible.")
				os.Exit(4)
			}
			return err
		}

		if outputFormat == "human" {
			stats := engine.CompletionStats{
				FilesCreated:  totalFiles,
				FilesModified: 0,
				FilesSkipped:  len(actions) - totalFiles,
			}
			
			completionCtx := make(map[string]any)
			for k, v := range ctx.Values {
				completionCtx[k] = v
			}
			completionCtx["OUTPUT_DIR"] = cwd
			if _, ok := ctx.Values["PROJECT_NAME"]; ok {
				completionCtx["PROJECT_NAME"] = ctx.Values["PROJECT_NAME"]
			}
			
			fmt.Print(tui.RenderCompletionWithContext(stats, bp, completionCtx))
		}

		return nil
	},
}

func checkLocal(s string) bool {
	return len(s) > 0 && (s[0] == '.' || s[0] == '/' || (len(s) > 1 && s[1] == ':'))
}

func init() {
	regenCmd.Flags().BoolVar(&regenDryRun, "dry-run", false, "Hanya simulasikan tanpa merubah file fisik")
	regenCmd.Flags().BoolVarP(&regenYes, "yes", "y", false, "Skip konfirmasi di tahap pra-penulisan")
	regenCmd.Flags().BoolVar(&regenNoHooks, "no-hooks", false, "Jangan jalankan action shell hook pada blueprint templatnya")
	rootCmd.AddCommand(regenCmd)
}
