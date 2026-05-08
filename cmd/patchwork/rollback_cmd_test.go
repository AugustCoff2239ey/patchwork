package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

func writeRollbackHistory(t *testing.T, path string, entries []history.Entry) {
	t.Helper()
	log := history.Log{Entries: entries}
	b, _ := json.Marshal(log)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("writeRollbackHistory: %v", err)
	}
}

func writeRollbackSnap(t *testing.T, path string, data map[string]string) {
	t.Helper()
	snap := snapshot.Snapshot{Data: data}
	b, _ := json.Marshal(snap)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("writeRollbackSnap: %v", err)
	}
}

func TestRunRollback_Success(t *testing.T) {
	dir := t.TempDir()
	snap1 := filepath.Join(dir, "snap1.json")
	snap2 := filepath.Join(dir, "snap2.json")
	writeRollbackSnap(t, snap1, map[string]string{"version": "1"})
	writeRollbackSnap(t, snap2, map[string]string{"version": "2"})

	now := time.Now().UTC()
	histPath := filepath.Join(dir, "history.json")
	writeRollbackHistory(t, histPath, []history.Entry{
		{Environment: "prod", SnapshotPath: snap1, CapturedAt: now.Add(-2 * time.Hour)},
		{Environment: "prod", SnapshotPath: snap2, CapturedAt: now.Add(-1 * time.Hour)},
	})

	destPath := filepath.Join(dir, "result.json")
	if err := runRollback("prod", histPath, destPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("expected result snapshot to exist")
	}
}

func TestRunRollback_MissingHistory(t *testing.T) {
	dir := t.TempDir()
	err := runRollback("prod", filepath.Join(dir, "missing.json"), filepath.Join(dir, "out.json"))
	if err == nil {
		t.Fatal("expected error for missing history file")
	}
}

func TestRunRollback_EmptyEnvironment(t *testing.T) {
	dir := t.TempDir()
	err := runRollback("", filepath.Join(dir, "h.json"), filepath.Join(dir, "out.json"))
	if err == nil {
		t.Fatal("expected error for empty environment")
	}
}

func TestRunRollback_InsufficientHistory(t *testing.T) {
	dir := t.TempDir()
	histPath := filepath.Join(dir, "history.json")
	writeRollbackHistory(t, histPath, []history.Entry{
		{Environment: "staging", SnapshotPath: "/snap.json", CapturedAt: time.Now()},
	})
	err := runRollback("staging", histPath, filepath.Join(dir, "out.json"))
	if err == nil {
		t.Fatal("expected error for insufficient history entries")
	}
}
