package prune

import (
	"fmt"
	"time"

	"github.com/user/patchwork/internal/history"
)

// Options controls what history entries to prune.
type Options struct {
	// OlderThan removes entries older than this duration.
	OlderThan time.Duration
	// KeepLast retains at least this many entries per environment, regardless of age.
	KeepLast int
	// Environment restricts pruning to a specific environment (empty = all).
	Environment string
}

// Result summarises what was removed.
type Result struct {
	Removed int
	Retained int
}

// Apply removes history entries that match the prune criteria and returns the
// filtered log together with a summary of what was pruned.
func Apply(log history.Log, opts Options) (history.Log, Result, error) {
	if opts.KeepLast < 0 {
		return nil, Result{}, fmt.Errorf("prune: KeepLast must be >= 0")
	}

	cutoff := time.Now().Add(-opts.OlderThan)

	// Group entries by environment to honour KeepLast per env.
	byEnv := make(map[string][]history.Entry)
	for _, e := range log {
		byEnv[e.Environment] = append(byEnv[e.Environment], e)
	}

	kept := make(history.Log, 0, len(log))
	var removed int

	for env, entries := range byEnv {
		if opts.Environment != "" && env != opts.Environment {
			// Not targeted — keep everything.
			kept = append(kept, entries...)
			continue
		}

		// Entries are assumed newest-first after history.Log sorting.
		for i, e := range entries {
			withinKeepLast := opts.KeepLast > 0 && i < opts.KeepLast
			tooNew := opts.OlderThan > 0 && e.Timestamp.After(cutoff)

			if withinKeepLast || tooNew {
				kept = append(kept, e)
			} else {
				removed++
			}
		}
	}

	return kept, Result{Removed: removed, Retained: len(kept)}, nil
}
