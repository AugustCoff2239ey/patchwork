package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/patchwork/internal/export"
	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

// runExport generates a structured export of the diff history and writes it
// to the specified output file in the requested format (json or csv).
func runExport(historyFile, outputFile, format, env string) error {
	log, err := history.LoadLog(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no history found at %s: run 'patchwork capture' first", historyFile)
		}
		return fmt.Errorf("loading history: %w", err)
	}

	entries := log.Entries
	if env != "" {
		entries = filterEntriesByEnv(entries, env)
		if len(entries) == 0 {
			return fmt.Errorf("no history entries found for environment %q", env)
		}
	}

	// Build export records from history entries, loading snapshot pairs for each.
	var records []export.Record
	for i, entry := range entries {
		var prev *snapshot.Snapshot
		if i > 0 {
			prev, err = loadSnapshotForExport(entries[i-1].SnapshotPath)
			if err != nil {
				// Non-fatal: skip diff for this entry if previous snapshot is unavailable.
				fmt.Fprintf(os.Stderr, "warning: could not load previous snapshot for %s: %v\n", entry.SnapshotPath, err)
			}
		}

		curr, err := loadSnapshotForExport(entry.SnapshotPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load snapshot %s: %v\n", entry.SnapshotPath, err)
			continue
		}

		rec := export.Build(entry, prev, curr)
		records = append(records, rec)
	}

	if len(records) == 0 {
		return fmt.Errorf("no exportable records found")
	}

	// Ensure output directory exists.
	if dir := filepath.Dir(outputFile); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	if err := export.Write(records, outputFile, format); err != nil {
		return fmt.Errorf("writing export: %w", err)
	}

	fmt.Printf("Exported %d record(s) to %s (format: %s)\n", len(records), outputFile, format)
	return nil
}

// loadSnapshotForExport loads a snapshot from disk, returning nil without error
// if the path is empty (indicating no snapshot is associated).
func loadSnapshotForExport(path string) (*snapshot.Snapshot, error) {
	if path == "" {
		return nil, nil
	}
	return snapshot.Load(path)
}

// filterEntriesByEnv returns only the history entries matching the given environment.
func filterEntriesByEnv(entries []history.Entry, env string) []history.Entry {
	var filtered []history.Entry
	for _, e := range entries {
		if e.Environment == env {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
