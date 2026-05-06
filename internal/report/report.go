package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
)

// Summary holds aggregated drift statistics for a report.
type Summary struct {
	GeneratedAt time.Time
	TotalSnapshots int
	TotalChanges int
	AddedKeys int
	RemovedKeys int
	ModifiedKeys int
	Entries []ReportEntry
}

// ReportEntry pairs a history log entry with its computed diff changes.
type ReportEntry struct {
	Log history.LogEntry
	Changes []diff.Change
}

// Build constructs a Summary from a slice of log entries and their associated diffs.
func Build(entries []ReportEntry) Summary {
	s := Summary{
		GeneratedAt: time.Now().UTC(),
		TotalSnapshots: len(entries),
		Entries: entries,
	}
	for _, e := range entries {
		for _, c := range e.Changes {
			s.TotalChanges++
			switch c.Type {
			case diff.Added:
				s.AddedKeys++
			case diff.Removed:
				s.RemovedKeys++
			case diff.Modified:
				s.ModifiedKeys++
			}
		}
	}
	return s
}

// Render writes a human-readable report to the given writer.
func Render(w io.Writer, s Summary) error {
	fmt.Fprintf(w, "Patchwork Drift Report\n")
	fmt.Fprintf(w, "Generated: %s\n", s.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 40))
	fmt.Fprintf(w, "Snapshots reviewed : %d\n", s.TotalSnapshots)
	fmt.Fprintf(w, "Total changes      : %d\n", s.TotalChanges)
	fmt.Fprintf(w, "  Added keys       : %d\n", s.AddedKeys)
	fmt.Fprintf(w, "  Removed keys     : %d\n", s.RemovedKeys)
	fmt.Fprintf(w, "  Modified keys    : %d\n", s.ModifiedKeys)
	if len(s.Entries) > 0 {
		fmt.Fprintf(w, "\nPer-snapshot detail:\n")
		for _, e := range s.Entries {
			fmt.Fprintf(w, "  [%s] %s — %d change(s)\n",
				e.Log.Timestamp.Format("2006-01-02 15:04:05"),
				e.Log.File,
				len(e.Changes),
			)
		}
	}
	return nil
}
