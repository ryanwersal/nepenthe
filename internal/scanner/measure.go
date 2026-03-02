package scanner

import (
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

func measureConcurrency() int {
	return max(1, runtime.NumCPU()/4)
}

func MeasureSizes(ctx context.Context, results []ScanResult) []ScanResult {
	concurrency := measureConcurrency()
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := range results {
		g.Go(func() error {
			size, _ := measureDU(ctx, results[i].Path)
			results[i].SizeBytes = size

			count, _ := measureFileCount(results[i].Path)
			results[i].FileCount = count
			return nil
		})
	}
	_ = g.Wait()
	return results
}

// SizeMeasurement is the result of measuring a single directory.
type SizeMeasurement struct {
	Index     int
	SizeBytes int64
	FileCount int64
}

// MeasureSizesStream measures sizes concurrently and calls onMeasured for each
// result as it completes. This allows progressive UI updates.
func MeasureSizesStream(ctx context.Context, results []ScanResult, onMeasured func(SizeMeasurement)) {
	concurrency := measureConcurrency()
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := range results {
		g.Go(func() error {
			size, _ := measureDU(ctx, results[i].Path)
			count, _ := measureFileCount(results[i].Path)
			onMeasured(SizeMeasurement{
				Index:     i,
				SizeBytes: size,
				FileCount: count,
			})
			return nil
		})
	}
	_ = g.Wait()
}

func measureDU(ctx context.Context, path string) (int64, error) {
	out, err := exec.CommandContext(ctx, "du", "-sk", path).Output()
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0, nil
	}
	kb, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, err
	}
	return kb * 1024, nil
}

func measureFileCount(path string) (int64, error) {
	var count int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries we can't read
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}
