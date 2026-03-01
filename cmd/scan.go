package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/consent"
	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
	"github.com/spf13/cobra"
)

var (
	scanDryRun bool
	scanSizes  bool
	scanAll    bool
	scanQuiet  bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for excludable directories and apply exclusions",
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().BoolVar(&scanDryRun, "dry-run", false, "Preview only, no changes")
	scanCmd.Flags().BoolVar(&scanSizes, "sizes", false, "Measure directory sizes (slower)")
	scanCmd.Flags().BoolVar(&scanAll, "all", false, "Skip consent prompts")
	scanCmd.Flags().BoolVarP(&scanQuiet, "quiet", "q", false, "Suppress output except errors")
}

func runScan(cmd *cobra.Command, args []string) error {
	if err := tmutil.AssertAvailable(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Build rules: built-in + custom
	rules := scanner.BuildSentinelRules(customSentinelRules(cfg))

	// Convert custom fixed paths
	var customFixed []scanner.FixedPathRule
	for _, cf := range cfg.CustomFixedPaths {
		customFixed = append(customFixed, scanner.FixedPathRule{
			Path:      cf.Path,
			Ecosystem: cf.Ecosystem,
			Tier:      0,
		})
	}

	if !scanQuiet {
		fmt.Println("Scanning...")
	}

	// Run sentinel scan
	sentinelResults := scanner.ScanSentinelRules(scanner.WalkOptions{
		Roots:       cfg.Roots,
		Rules:       rules,
		Concurrency: cfg.Concurrency.ScanWorkers,
	})

	// Run fixed-path scan
	fixedResults, err := scanner.ScanFixedPaths(cfg.EnabledTiers, customFixed)
	if err != nil {
		return fmt.Errorf("scanning fixed paths: %w", err)
	}

	// Combine results
	results := append(sentinelResults, fixedResults...)

	// Filter by consent
	if !scanAll {
		var filtered []scanner.ScanResult
		for _, r := range results {
			if r.Tier <= 1 {
				filtered = append(filtered, r)
				continue
			}
			ok, err := consent.CheckTierConsent(r.Tier, cfg, os.Stdout)
			if err != nil {
				return err
			}
			if ok {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Measure sizes if requested
	if scanSizes || scanDryRun {
		if !scanQuiet {
			fmt.Println("Measuring sizes...")
		}
		results = scanner.MeasureSizes(results, cfg.Concurrency.MeasureWorkers)
	}

	// Dry run: print table and exit
	if scanDryRun {
		printScanTable(results)
		return nil
	}

	// Apply exclusions
	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	var applied, failed int
	for _, r := range results {
		if r.IsExcluded {
			continue
		}
		if err := tmutil.AddExclusion(r.Path); err != nil {
			failed++
			if !scanQuiet {
				fmt.Fprintf(os.Stderr, "Failed to exclude: %s\n", r.Path)
			}
			continue
		}
		state.AddExclusion(&st, r.Path, r.Tier, r.Type, r.Ecosystem)
		applied++
	}

	if err := state.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if !scanQuiet {
		fmt.Printf("Applied %d exclusions", applied)
		if failed > 0 {
			fmt.Printf(" (%d failed)", failed)
		}
		fmt.Println()
	}

	return nil
}

func printScanTable(results []scanner.ScanResult) {
	if len(results) == 0 {
		fmt.Println("No directories found.")
		return
	}

	// Header
	fmt.Printf("%-50s %-24s %10s %10s\n", "Path", "Ecosystem", "Size", "Files")
	fmt.Println(strings.Repeat("─", 98))

	var totalSize int64
	var totalFiles int64
	for _, r := range results {
		path := r.Path
		if len(path) > 48 {
			path = "…" + path[len(path)-47:]
		}
		sizeStr := ""
		filesStr := ""
		if r.SizeBytes > 0 {
			sizeStr = format.Bytes(r.SizeBytes)
			totalSize += r.SizeBytes
		}
		if r.FileCount > 0 {
			filesStr = format.Count(r.FileCount)
			totalFiles += r.FileCount
		}
		fmt.Printf("%-50s %-24s %10s %10s\n", path, r.Ecosystem, sizeStr, filesStr)
	}

	fmt.Println(strings.Repeat("─", 98))
	fmt.Printf("Total: %d directories %44s %10s\n",
		len(results), format.Bytes(totalSize), format.Count(totalFiles))
}
