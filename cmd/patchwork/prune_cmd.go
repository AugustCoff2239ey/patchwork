package main

import (
	"fmt"
	"os"
	"time"

	"github.com/user/patchwork/internal/history"
	"github.com/user/patchwork/internal/prune"
)

// runPrune removes old history entries according to the supplied flags.
//
// Usage: patchwork prune --older-than <duration> [--keep-last <n>] [--env <name>] <history-file>
func runPrune(historyFile string, olderThan time.Duration, keepLast int, env string, dryRun bool) error {
	log, err := history.LoadLog(historyFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("prune: load history: %w", err)
	}

	if len(log) == 0 {
		fmt.Println("prune: no history entries found — nothing to do")
		return nil
	}

	opts := prune.Options{
		OlderThan:   olderThan,
		KeepLast:    keepLast,
		Environment: env,
	}

	kept, result, err := prune.Apply(log, opts)
	if err != nil {
		return fmt.Errorf("prune: %w", err)
	}

	if dryRun {
		fmt.Printf("prune (dry-run): would remove %d entries, retain %d\n",
			result.Removed, result.Retained)
		return nil
	}

	if result.Removed == 0 {
		fmt.Println("prune: no entries matched the prune criteria")
		return nil
	}

	if err := history.SaveLog(historyFile, kept); err != nil {
		return fmt.Errorf("prune: save history: %w", err)
	}

	fmt.Printf("prune: removed %d entries, %d retained\n", result.Removed, result.Retained)
	return nil
}
