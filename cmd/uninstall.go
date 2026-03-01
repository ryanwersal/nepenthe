package cmd

import (
	"fmt"

	"github.com/ryanwersal/nepenthe/internal/launchd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove launchd scheduled scan",
	RunE:  runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	if err := launchd.Uninstall(); err != nil {
		return err
	}
	fmt.Println("Launchd plist removed.")
	return nil
}
