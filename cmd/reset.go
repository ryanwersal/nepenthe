package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
	"github.com/spf13/cobra"
)

var (
	resetDryRun bool
	resetForce  bool
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remove all exclusions tracked by Nepenthe",
	RunE:  runReset,
}

func init() {
	resetCmd.Flags().BoolVar(&resetDryRun, "dry-run", false, "Preview only")
	resetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Skip confirmation prompt")
}

func runReset(cmd *cobra.Command, args []string) error {
	if err := tmutil.AssertAvailable(); err != nil {
		return err
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if len(st.Exclusions) == 0 {
		fmt.Println("No exclusions tracked by Nepenthe.")
		return nil
	}

	fmt.Printf("Exclusions to remove (%d):\n", len(st.Exclusions))
	for _, e := range st.Exclusions {
		fmt.Printf("  %s\n", e.Path)
	}

	if resetDryRun {
		return nil
	}

	if !resetForce {
		fmt.Printf("\nRemove %d exclusions? [y/N] ", len(st.Exclusions))
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if len(line) == 0 || (line[0] != 'y' && line[0] != 'Y') {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	var removed, stale int64
	var failed int64
	exclusions := state.ClearAll(&st)

	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, 16)
	)

	for _, e := range exclusions {
		if _, err := os.Stat(e.Path); os.IsNotExist(err) {
			atomic.AddInt64(&stale, 1)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(e state.TrackedExclusion) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := tmutil.RemoveExclusion(e.Path); err != nil {
				atomic.AddInt64(&failed, 1)
				mu.Lock()
				state.AddExclusion(&st, e.Path, e.Category, e.Type, e.Ecosystem)
				mu.Unlock()
				slog.Warn("removal failed", "path", e.Path, "err", err)
				return
			}
			atomic.AddInt64(&removed, 1)
		}(e)
	}
	wg.Wait()

	if err := state.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Printf("Removed %d exclusions", removed)
	if stale > 0 {
		fmt.Printf(", %d stale", stale)
	}
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()

	return nil
}
