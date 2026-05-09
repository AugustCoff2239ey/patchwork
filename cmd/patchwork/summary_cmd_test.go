package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/patchwork/internal/history"
	"github.com/user/patchwork/internal/snapshot"
)

func writeSummaryHistory(t *testing.T, dir string, entries []history.Entry) string {
	t.Helper()
	path := filepath.Join(dir, "history.json")
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write history: %v", err)
	}
	return path
}

func makeSummarySnapshot(t *testing.T, dir, env, name string, data map[string]string) string {
	t.Helper()
	snap := snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
	path := filepath.Join(dir, name+".json")
	if err := snapshot.Save(snap, path); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	return path
}

func TestRunSummary_NoHistory(t *testing.T) {
	dir := t.TempDir()
	histPath := filepath.Join(dir, "history.json")

	err := runSummary(histPath, "", false)
	if err == nil {
		t.Fatal("expected error for missing history file, got nil")
	}
}

func TestRunSummary_EmptyHistory(t *testing.T) {
	dir := t.TempDir()
	histPath := writeSummaryHistory(t, dir, []history.Entry{})

	// Empty history should succeed but produce minimal output
	err := runSummary(histPath, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSummary_WithEntries(t *testing.T) {
	dir := t.TempDir()

	snap1 := makeSummarySnapshot(t, dir, "production", "snap1", map[string]string{
		"db.host": "prod-db",
		"cache.ttl": "300",
	})
	snap2 := makeSummarySnapshot(t, dir, "production", "snap2", map[string]string{
		"db.host": "prod-db-2",
		"cache.ttl": "600",
	})

	entries := []history.Entry{
		{
			Environment:  "production",
			Timestamp:    time.Now().Add(-2 * time.Hour),
			SnapshotPath: snap1,
			ChangeCount:  2,
		},
		{
			Environment:  "production",
			Timestamp:    time.Now().Add(-1 * time.Hour),
			SnapshotPath: snap2,
			ChangeCount:  1,
		},
	}
	histPath := writeSummaryHistory(t, dir, entries)

	err := runSummary(histPath, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSummary_FilterByEnvironment(t *testing.T) {
	dir := t.TempDir()

	snap1 := makeSummarySnapshot(t, dir, "staging", "snap1", map[string]string{"key": "val"})

	entries := []history.Entry{
		{
			Environment:  "staging",
			Timestamp:    time.Now().Add(-1 * time.Hour),
			SnapshotPath: snap1,
			ChangeCount:  1,
		},
		{
			Environment:  "production",
			Timestamp:    time.Now().Add(-30 * time.Minute),
			SnapshotPath: snap1,
			ChangeCount:  3,
		},
	}
	histPath := writeSummaryHistory(t, dir, entries)

	// Filter to staging only — should not error
	err := runSummary(histPath, "staging", false)
	if err != nil {
		t.Fatalf("unexpected error with env filter: %v", err)
	}
}

func TestRunSummary_VerboseFlag(t *testing.T) {
	dir := t.TempDir()

	snap1 := makeSummarySnapshot(t, dir, "production", "snap1", map[string]string{
		"app.version": "1.0",
	})

	entries := []history.Entry{
		{
			Environment:  "production",
			Timestamp:    time.Now().Add(-1 * time.Hour),
			SnapshotPath: snap1,
			ChangeCount:  1,
		},
	}
	histPath := writeSummaryHistory(t, dir, entries)

	err := runSummary(histPath, "", true)
	if err != nil {
		t.Fatalf("unexpected error in verbose mode: %v", err)
	}
}
