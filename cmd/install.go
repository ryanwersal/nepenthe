package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/launchd"
	"github.com/spf13/cobra"
)

var installInterval int

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install launchd scheduled scan",
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().IntVar(&installInterval, "interval", 0, "Interval in seconds (default: from config or 86400)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	interval := installInterval
	if interval == 0 {
		interval = cfg.Schedule.IntervalSeconds
	}

	// Resolve binary path
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return fmt.Errorf("resolving binary symlinks: %w", err)
	}

	if err := launchd.Install(binary, interval); err != nil {
		return err
	}

	plistPath, err := launchd.PlistFilePath()
	if err != nil {
		return err
	}
	fmt.Printf("Installed launchd plist: %s\n", plistPath)
	fmt.Printf("Scan interval: %s (%d seconds)\n", format.Interval(interval), interval)
	return nil
}
