package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/patchwork/internal/history"
)

const defaultHistoryPath = ".patchwork/history.json"

// runHistory prints the recorded snapshot history, optionally filtered by file.
func runHistory(args []string) error {
	log, err := history.LoadLog(defaultHistoryPath)
	if err != nil {
		return fmt.Errorf("could not load history: %w", err)
	}

	entries := log.Sorted()

	if len(args) > 0 {
		filterPath := args[0]
		filtered := entries[:0]
		for _, e := range entries {
			if e.FilePath == filterPath {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if len(entries) == 0 {
		fmt.Println("No history recorded yet.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tLABEL\tFILE\tSNAPSHOT")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.Label,
			e.FilePath,
			e.SnapshotPath,
		)
	}
	return w.Flush()
}

// recordHistory appends a new entry to the on-disk history log.
func recordHistory(label, filePath, snapshotPath string) error {
	log, err := history.LoadLog(defaultHistoryPath)
	if err != nil {
		return fmt.Errorf("recordHistory: load: %w", err)
	}
	log.Add(label, filePath, snapshotPath)
	if err := history.SaveLog(defaultHistoryPath, log); err != nil {
		return fmt.Errorf("recordHistory: save: %w", err)
	}
	return nil
}
