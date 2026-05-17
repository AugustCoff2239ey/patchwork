package truncate

import (
	"fmt"
	"sort"
	"time"

	"github.com/patchwork/internal/history"
)

// Options controls how truncation is applied.
type Options struct {
	Environment string
	Before      time.Time
	MaxEntries  int
}

// Result summarises what was removed.
type Result struct {
	Removed     int
	Remaining   int
	Environment string
}

// Apply removes history entries that match the given options.
// At least one of Before or MaxEntries must be set.
func Apply(log history.Log, opts Options) (history.Log, Result, error) {
	if opts.Before.IsZero() && opts.MaxEntries <= 0 {
		return log, Result{}, fmt.Errorf("truncate: at least one of --before or --max must be provided")
	}

	entries := log.Entries
	if opts.Environment != "" {
		var filtered []history.Entry
		for _, e := range entries {
			if e.Environment == opts.Environment {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Sort ascending by timestamp so we can trim oldest first.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	var keep []history.Entry
	for _, e := range entries {
		if !opts.Before.IsZero() && e.Timestamp.Before(opts.Before) {
			continue
		}
		keep = append(keep, e)
	}

	if opts.MaxEntries > 0 && len(keep) > opts.MaxEntries {
		keep = keep[len(keep)-opts.MaxEntries:]
	}

	removed := len(entries) - len(keep)

	// Re-integrate non-targeted entries when an environment filter is active.
	if opts.Environment != "" {
		var merged []history.Entry
		keptSet := make(map[string]bool, len(keep))
		for _, e := range keep {
			keptSet[e.SnapshotPath] = true
		}
		for _, e := range log.Entries {
			if e.Environment != opts.Environment || keptSet[e.SnapshotPath] {
				merged = append(merged, e)
			}
		}
		keep = merged
	}

	return history.Log{Entries: keep}, Result{
		Removed:     removed,
		Remaining:   len(keep),
		Environment: opts.Environment,
	}, nil
}

// Render returns a human-readable summary of a truncation result.
func Render(r Result) string {
	env := r.Environment
	if env == "" {
		env = "all"
	}
	return fmt.Sprintf("truncate: removed %d entr%s, %d remaining (environment: %s)\n",
		r.Removed, plural(r.Removed), r.Remaining, env)
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
