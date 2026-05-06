package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/report"
	"github.com/patchwork/internal/snapshot"
)

// runReport generates a drift summary report across all recorded history entries.
func runReport(historyFile string) error {
	log, err := history.LoadLog(historyFile)
	if err != nil {
		return fmt.Errorf("loading history: %w", err)
	}
	if len(log.Entries) == 0 {
		fmt.Println("No history entries found. Capture some snapshots first.")
		return nil
	}

	var entries []report.ReportEntry
	for i, entry := range log.Entries {
		var changes []diff.Change
		if i > 0 {
			prev, err := loadSnapshotForEntry(log.Entries[i-1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: skipping entry %s: %v\n", entry.File, err)
				continue
			}
			curr, err := loadSnapshotForEntry(entry)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: skipping entry %s: %v\n", entry.File, err)
				continue
			}
			changes = diff.Compare(prev, curr)
		}
		entries = append(entries, report.ReportEntry{Log: entry, Changes: changes})
	}

	summary := report.Build(entries)
	return report.Render(os.Stdout, summary)
}

func loadSnapshotForEntry(e history.LogEntry) (snapshot.Snapshot, error) {
	snapshotPath := filepath.Join(".patchwork", "snapshots", e.SnapshotID+".json")
	return snapshot.Load(snapshotPath)
}
