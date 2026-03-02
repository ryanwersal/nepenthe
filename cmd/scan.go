package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/consent"
	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	scanDryRun bool
	scanSizes  bool
	scanAll    bool
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
}

func runScan(cmd *cobra.Command, args []string) error {
	scanStart := time.Now()

	if err := tmutil.AssertAvailable(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Build rules
	rules := scanner.BuildSentinelRules()

	// Convert custom fixed paths
	var customFixed []scanner.FixedPathRule
	for _, cf := range cfg.CustomFixedPaths {
		customFixed = append(customFixed, scanner.FixedPathRule{
			Path:      cf.Path,
			Ecosystem: cf.Ecosystem,
			Category:  scanner.CategoryCustom,
		})
	}

	slog.Info("config loaded",
		"roots", len(cfg.Roots),
		"rules", len(rules),
		"categories", cfg.EnabledCategories,
	)

	ctx := cmd.Context()
	slog.Info("scan started")

	// Run sentinel scan
	sentinelResults := scanner.ScanSentinelRules(ctx, scanner.WalkOptions{
		Roots: cfg.Roots,
		Rules: rules,
	})

	for _, r := range sentinelResults {
		slog.Debug("sentinel match", "path", r.Path, "ecosystem", r.Ecosystem)
	}

	// Run fixed-path scan
	fixedResults, err := scanner.ScanFixedPaths(ctx, enabledCategories(cfg), customFixed)
	if err != nil {
		return fmt.Errorf("scanning fixed paths: %w", err)
	}

	for _, r := range fixedResults {
		slog.Debug("fixed path match", "path", r.Path, "ecosystem", r.Ecosystem)
	}

	// Combine results
	results := make([]scanner.ScanResult, 0, len(sentinelResults)+len(fixedResults))
	results = append(results, sentinelResults...)
	results = append(results, fixedResults...)

	slog.Info("scan complete", "found", len(results), "elapsed", time.Since(scanStart))

	// Filter by consent
	if !scanAll {
		var filtered []scanner.ScanResult
		for _, r := range results {
			if r.Category == scanner.CategoryDependencies || r.Category == scanner.CategoryCustom {
				filtered = append(filtered, r)
				continue
			}
			ok, err := consent.CheckCategoryConsent(string(r.Category), cfg, os.Stdout)
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
		slog.Info("measuring sizes")
		results = scanner.MeasureSizes(ctx, results)
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

	type exclusionResult struct {
		path      string
		category  string
		typ       string
		ecosystem string
	}

	var (
		skipped int
		mu      sync.Mutex
	)
	var successes []exclusionResult
	var failed int

	slog.Info("applying exclusions", "count", len(results))

	g := new(errgroup.Group)
	g.SetLimit(16)

	for _, r := range results {
		if r.IsExcluded {
			slog.Debug("exclusion skipped", "path", r.Path, "reason", "already excluded")
			skipped++
			continue
		}
		g.Go(func() error {
			if err := tmutil.AddExclusion(r.Path); err != nil {
				mu.Lock()
				failed++
				mu.Unlock()
				slog.Warn("exclusion failed", "path", r.Path, "err", err)
				return nil
			}
			slog.Info("exclusion applied", "path", r.Path, "category", r.Category, "ecosystem", r.Ecosystem)
			mu.Lock()
			successes = append(successes, exclusionResult{r.Path, string(r.Category), r.Type, r.Ecosystem})
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()

	for _, s := range successes {
		state.AddExclusion(&st, s.path, s.category, s.typ, s.ecosystem)
	}

	if err := state.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	applied := len(successes)
	slog.Info("scan finished",
		"applied", applied,
		"failed", failed,
		"skipped", skipped,
		"elapsed", time.Since(scanStart),
	)

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
