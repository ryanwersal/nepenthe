package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	logVerbose bool
	logQuiet   bool
	logFormat  string
)

var rootCmd = &cobra.Command{
	Use:   "nepenthe",
	Short: "Selective forgetfulness for macOS Time Machine",
	Long:  "Nepenthe manages Time Machine exclusions for build artifacts, caches, and other regenerable directories.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := slog.LevelInfo
		if logVerbose {
			level = slog.LevelDebug
		} else if logQuiet {
			level = slog.LevelWarn
		}

		opts := &slog.HandlerOptions{Level: level}
		var handler slog.Handler
		switch logFormat {
		case "json":
			handler = slog.NewJSONHandler(os.Stderr, opts)
		case "text":
			handler = slog.NewTextHandler(os.Stderr, opts)
		default:
			return fmt.Errorf("invalid --log-format %q: must be text or json", logFormat)
		}
		slog.SetDefault(slog.New(handler))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&logVerbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&logQuiet, "quiet", "q", false, "Only show warnings and errors")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format: text or json")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func enabledCategories(cfg config.Config) []scanner.Category {
	cats := make([]scanner.Category, len(cfg.EnabledCategories))
	for i, c := range cfg.EnabledCategories {
		cats[i] = scanner.Category(c)
	}
	return cats
}
