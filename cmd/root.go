package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/Reivhell/symphony/internal/cliupdate"
	"github.com/Reivhell/symphony/internal/tui"
)

var Version = "dev"
var (
	cfgFile      string
	verbose      bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "symphony",
	Short: "Symphony — The Adaptive Scaffolding Engine",
	Long: `Symphony is an adaptive scaffolding engine that orchestrates project
generation based on a data-driven template system.

It reads a template.yaml blueprint, interactively prompts the user for
configuration values, and generates a fully customized project structure.`,
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if errors.Is(err, tui.ErrUserCancelled) {
			fmt.Println(tui.StyleMuted.Render("\n  Dibatalkan."))
			os.Exit(4)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !verbose {
			return
		}
		if cmd != nil && cmd.Name() == "version" {
			return
		}
		go func() {
			cliupdate.MaybePrintNewerRelease(Version)
		}()
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Tampilkan log detail")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path ke config file kustom")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "human", "Format output: human | json")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		viper.AddConfigPath(home + "/.symphony")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}
