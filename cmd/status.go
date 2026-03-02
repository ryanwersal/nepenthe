package cmd

import (
	"fmt"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show exclusion status dashboard",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	if err := tmutil.AssertAvailable(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	rules := scanner.BuildSentinelRules(customSentinelRules(cfg))

	fmt.Println("Scanning...")
	sentinelResults := scanner.ScanSentinelRules(scanner.WalkOptions{
		Roots: cfg.Roots,
		Rules: rules,
	})
	fixedResults, err := scanner.ScanFixedPaths(enabledCategories(cfg), nil)
	if err != nil {
		return fmt.Errorf("scanning fixed paths: %w", err)
	}
	results := append(sentinelResults, fixedResults...)

	var byNepenthe, byOther, notExcluded []scanner.ScanResult
	for _, r := range results {
		if r.IsExcluded {
			if state.IsTracked(&st, r.Path) {
				byNepenthe = append(byNepenthe, r)
			} else {
				byOther = append(byOther, r)
			}
		} else {
			notExcluded = append(notExcluded, r)
		}
	}

	fmt.Printf("\nExcluded by Nepenthe:  %d directories\n", len(byNepenthe))
	fmt.Printf("Excluded by other:     %d directories\n", len(byOther))
	fmt.Printf("Not excluded:          %d directories\n", len(notExcluded))
	fmt.Printf("Total:                 %d directories found\n", len(results))

	if len(byNepenthe) > 0 {
		fmt.Println("\n  Excluded by Nepenthe:")
		for _, r := range byNepenthe {
			fmt.Printf("    %-20s %s\n", r.Ecosystem, r.Path)
		}
	}
	if len(byOther) > 0 {
		fmt.Println("\n  Excluded by other:")
		for _, r := range byOther {
			fmt.Printf("    %-20s %s\n", r.Ecosystem, r.Path)
		}
	}
	if len(notExcluded) > 0 {
		fmt.Println("\n  Not excluded:")
		for _, r := range notExcluded {
			fmt.Printf("    %-20s %s\n", r.Ecosystem, r.Path)
		}
	}

	return nil
}
