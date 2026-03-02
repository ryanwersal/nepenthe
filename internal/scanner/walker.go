package scanner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ryanwersal/nepenthe/internal/tmutil"
)

func ScanSentinelRules(ctx context.Context, opts WalkOptions) []ScanResult {
	concurrency := max(2, runtime.NumCPU()/2)

	// Build sentinel index: sentinel filename -> []SentinelRule
	sentinelIndex := make(map[string][]SentinelRule)
	for _, rule := range opts.Rules {
		for _, s := range rule.Sentinels {
			sentinelIndex[s] = append(sentinelIndex[s], rule)
		}
	}

	var (
		mu      sync.Mutex
		results []ScanResult
		seen    = make(map[string]bool)
		work    = make(chan string, concurrency*2)
		wg      sync.WaitGroup
	)

	addResult := func(r ScanResult) {
		mu.Lock()
		defer mu.Unlock()
		if seen[r.Path] {
			return
		}
		seen[r.Path] = true
		results = append(results, r)
		if opts.OnFound != nil {
			opts.OnFound(r)
		}
	}

	var processDir func(dir string)
	processDir = func(dir string) {
		if ctx.Err() != nil {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		// Separate files and dirs
		files := make(map[string]bool)
		var subdirs []string
		for _, e := range entries {
			if e.IsDir() {
				if !PruneDirs[e.Name()] {
					subdirs = append(subdirs, filepath.Join(dir, e.Name()))
				}
			} else {
				files[e.Name()] = true
			}
		}

		// Check sentinel files
		for filename := range files {
			rules, ok := sentinelIndex[filename]
			if !ok {
				continue
			}
			for _, rule := range rules {
				targetPath := filepath.Join(dir, rule.Directory)
				info, err := os.Stat(targetPath)
				if err != nil || !info.IsDir() {
					continue
				}
				excluded, _ := tmutil.IsExcluded(targetPath)
				addResult(ScanResult{
					Path:       targetPath,
					Ecosystem:  rule.Ecosystem,
					Category:   CategoryDependencies,
					Type:       "sticky",
					IsExcluded: excluded,
				})
			}
		}

		// Enqueue subdirectories
		for _, sub := range subdirs {
			wg.Add(1)
			// Non-blocking send; if channel is full, spawn inline to avoid deadlock
			select {
			case work <- sub:
			default:
				processDir(sub)
				wg.Done()
			}
		}
	}

	// Start fixed pool of workers
	for range concurrency {
		go func() {
			for dir := range work {
				processDir(dir)
				wg.Done()
			}
		}()
	}

	// Seed roots
	for _, root := range opts.Roots {
		wg.Add(1)
		work <- root
	}

	wg.Wait()
	close(work)
	return results
}

func ScanFixedPaths(ctx context.Context, enabledCategories []Category, customPaths []FixedPathRule) ([]ScanResult, error) {
	categories, err := FixedPathCategories()
	if err != nil {
		return nil, err
	}
	var rules []FixedPathRule

	for _, cat := range enabledCategories {
		if paths, ok := categories[cat]; ok {
			rules = append(rules, paths...)
		}
	}
	rules = append(rules, customPaths...)

	var (
		mu      sync.Mutex
		results []ScanResult
		wg      sync.WaitGroup
	)

	for _, rule := range rules {
		wg.Add(1)
		go func(r FixedPathRule) {
			defer wg.Done()
			_, err := os.ReadDir(r.Path)
			if err != nil {
				return
			}
			excluded, _ := tmutil.IsExcluded(r.Path)
			mu.Lock()
			results = append(results, ScanResult{
				Path:       r.Path,
				Ecosystem:  r.Ecosystem,
				Category:   r.Category,
				Type:       "sticky",
				IsExcluded: excluded,
			})
			mu.Unlock()
		}(rule)
	}

	wg.Wait()
	return results, nil
}
