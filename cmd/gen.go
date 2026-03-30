package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Reivhell/symphony/internal/blueprint"
	"github.com/Reivhell/symphony/internal/engine"
	"github.com/Reivhell/symphony/internal/remote"
	"github.com/Reivhell/symphony/internal/tui"
	"github.com/Reivhell/symphony/internal/version"
)

type fetchWrapper struct {
	cacheDir string
}

func (f fetchWrapper) Fetch(s string) (string, error) {
	path, _, err := remote.Fetch(s, f.cacheDir)
	return path, err
}

var (
	outDir  string
	dryRun  bool
	noHooks bool
	yesAll  bool
)

var genCmd = &cobra.Command{
	Use:   "gen <source>",
	Short: "Menjalankan scaffolding dari sebuah template source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		if outputFormat == "human" {
			fmt.Print(tui.RenderHeader(version.Version))
		}

		cacheDir, err := remote.CacheDir()
		if err != nil {
			return fmt.Errorf("gagal mendapatkan konfigurasi direktori cache: %w", err)
		}

		absSource, fetchMeta, err := remote.Fetch(source, cacheDir)
		if err != nil {
			return fmt.Errorf("gagal mengambil root template: %w", err)
		}

		bp, err := blueprint.Parse(absSource)
		if err != nil {
			return err
		}

		// Implementasi Resolve untuk Extends multilevel template
		wrapper := fetchWrapper{cacheDir: cacheDir}
		bp, err = blueprint.Resolve(bp, wrapper)
		if err != nil {
			return fmt.Errorf("gagal meresume inheritance template: %w", err)
		}

		// Tambahkan meta ke blueprint runtime execution jika ada (untuk commit/version trace)
		if fetchMeta.Commit != "" {
			bp.Version = fetchMeta.Version + "+" + fetchMeta.Commit
		}

		if outDir == "" {
			outDir, _ = os.Getwd()
		}
		absOutDir, err := filepath.Abs(outDir)
		if err != nil {
			return err
		}

		var answers map[string]any
		if yesAll {
			answers = make(map[string]any)
			for _, p := range bp.Prompts {
				answers[p.ID] = p.Default
			}
		} else {
			if outputFormat == "human" {
				fmt.Printf("  ◆ %s\n", tui.StyleHeader.Render("Konfigurasi Proyek Baru"))
				fmt.Printf("  %s\n\n", tui.Divider(54))
			}
			answers, err = tui.RunPrompts(bp)
			if err != nil {
				if errors.Is(err, tui.ErrUserCancelled) {
					fmt.Println(tui.StyleMuted.Render("\n  Dibatalkan."))
					os.Exit(4)
				}
				return err
			}
		}

		var ws *engine.WriteSession
		if !dryRun {
			ws = engine.NewWriteSession()
		}
		ctx := &engine.EngineContext{
			Values:       answers,
			SourceDir:    absSource,
			OutputDir:    absOutDir,
			DryRun:       dryRun,
			NoHooks:      noHooks,
			YesAll:       yesAll,
			Format:       outputFormat,
			Plugins:      engine.PluginRenderersFromBlueprint(bp),
			ExecCtx:      cmd.Context(),
			WriteSession: ws,
			Meta: engine.BlueprintMeta{
				Name:    bp.Name,
				Version: bp.Version,
				Source:  source,
				Commit:  "local",
			},
		}

		eng := engine.New(bp, ctx)
		actions, err := eng.Prepare()
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			tui.RenderJSONSummary(actions)
		} else {
			if !yesAll {
				cont, err := tui.RenderSummary(actions)
				if err != nil {
					if errors.Is(err, tui.ErrUserCancelled) {
						fmt.Println(tui.StyleMuted.Render("\n  Dibatalkan."))
						os.Exit(4)
					}
					return err
				}
				if !cont {
					fmt.Println(tui.StyleMuted.Render("\n  Dibatalkan."))
					os.Exit(4)
				}
			}
		}

		totalFiles := 0
		for _, a := range actions {
			if a.ShouldRun && a.Original.Type != "shell" {
				totalFiles++
			}
		}

		if outputFormat == "human" {
			ctx.Reporter = tui.NewTUIReporter(totalFiles)
		} else {
			ctx.Reporter = tui.NewJSONReporter(totalFiles)
		}

		if err := eng.Execute(cmd.Context(), actions); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Fprintln(os.Stderr, "Operation cancelled; rolled back partial output where possible.")
				os.Exit(4)
			}
			return err
		}

		if outputFormat == "human" {
			// Retrieve completion stats. Reporter handles the visual.
			stats := engine.CompletionStats{
				FilesCreated:  totalFiles,
				FilesModified: 0,
				FilesSkipped:  len(actions) - totalFiles,
			}
			// Sertakan OUTPUT_DIR dan semua jawaban agar completion_message bisa dirender
			completionCtx := make(map[string]any)
			for k, v := range answers {
				completionCtx[k] = v
			}
			completionCtx["OUTPUT_DIR"] = absOutDir
			completionCtx["PROJECT_NAME"] = answers["PROJECT_NAME"]
			fmt.Print(tui.RenderCompletionWithContext(stats, bp, completionCtx))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringVarP(&outDir, "out", "o", "", "Direktori output")
	genCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Preview file tanpa menulis ke disk")
	genCmd.Flags().BoolVar(&noHooks, "no-hooks", false, "Lewati execution hook")
	genCmd.Flags().BoolVarP(&yesAll, "yes", "y", false, "Lewati konfirmasi")
}
