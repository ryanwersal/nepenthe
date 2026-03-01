package cmd

import (
	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nepenthe",
	Short: "Selective forgetfulness for macOS Time Machine",
	Long:  "Nepenthe manages Time Machine exclusions for build artifacts, caches, and other regenerable directories.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
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

func customSentinelRules(cfg config.Config) []scanner.SentinelRule {
	custom := make([]scanner.SentinelRule, 0, len(cfg.CustomSentinelRules))
	for _, cr := range cfg.CustomSentinelRules {
		custom = append(custom, scanner.SentinelRule{
			Directory: cr.Directory,
			Sentinels: cr.Sentinels,
			Ecosystem: cr.Ecosystem,
		})
	}
	return custom
}
