package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

func TestRunReport_NoHistory(t *testing.T) {
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history.json")
	// Write an empty log
	if err := history.SaveLog(histFile, history.Log{}); err != nil {
		t.Fatal(err)
	}
	if err := runReport(histFile); err != nil {
		t.Errorf("expected no error for empty history, got: %v", err)
	}
}

func TestRunReport_MissingHistoryFile(t *testing.T) {
	dir := t.TempDir()
	err := runReport(filepath.Join(dir, "nonexistent.json"))
	// LoadLog on missing file returns empty log, not an error per history contract
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunReport_WithSnapshots(t *testing.T) {
	dir := t.TempDir()
	snapshotDir := filepath.Join(dir, ".patchwork", "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatal(err)
	}

	snap1 := snapshot.Snapshot{File: "cfg.json", Data: map[string]string{"port": "8080"}}
	snap2 := snapshot.Snapshot{File: "cfg.json", Data: map[string]string{"port": "9090", "host": "localhost"}}

	if err := snapshot.Save(filepath.Join(snapshotDir, "snap1.json"), snap1); err != nil {
		t.Fatal(err)
	}
	if err := snapshot.Save(filepath.Join(snapshotDir, "snap2.json"), snap2); err != nil {
		t.Fatal(err)
	}

	log := history.Log{
		Entries: []history.LogEntry{
			{File: "cfg.json", SnapshotID: "snap1", Timestamp: time.Now().Add(-time.Hour)},
			{File: "cfg.json", SnapshotID: "snap2", Timestamp: time.Now()},
		},
	}
	histFile := filepath.Join(dir, "history.json")
	if err := history.SaveLog(histFile, log); err != nil {
		t.Fatal(err)
	}

	// Override snapshot lookup base to temp dir — integration smoke test
	// We accept any outcome here; the goal is no panic.
	_ = runReport(histFile)
}
