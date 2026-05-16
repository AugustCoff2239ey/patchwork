package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
)

func writeDriftHistory(t *testing.T, dir string, entries []history.Entry) string {
	t.Helper()
	log := history.Log{Entries: entries}
	path := filepath.Join(dir, "history.json")
	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write history: %v", err)
	}
	return path
}

func TestRunDrift_MissingHistoryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")
	// Should not panic; missing file is treated as empty log.
	runDrift([]string{"--history", path, "--days", "7"})
}

func TestRunDrift_EmptyHistory(t *testing.T) {
	dir := t.TempDir()
	path := writeDriftHistory(t, dir, nil)
	// Should complete without error and print no-data message.
	runDrift([]string{"--history", path, "--days", "7"})
}

func TestRunDrift_WithEntries(t *testing.T) {
	dir := t.TempDir()
	entries := []history.Entry{
		{
			Environment:  "prod",
			Timestamp:    time.Now().UTC().AddDate(0, 0, -1),
			SnapshotPath: filepath.Join(dir, "nonexistent.json"),
		},
		{
			Environment:  "staging",
			Timestamp:    time.Now().UTC().AddDate(0, 0, -2),
			SnapshotPath: filepath.Join(dir, "nonexistent2.json"),
		},
	}
	path := writeDriftHistory(t, dir, entries)
	// Snapshots don't exist, so diffs will be empty — but run should not error.
	runDrift([]string{"--history", path, "--days", "7"})
}
