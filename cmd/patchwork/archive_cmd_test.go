package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
)

func writeArchiveHistory(t *testing.T, dir string, entries []history.Entry) string {
	t.Helper()
	path := filepath.Join(dir, "history.json")
	log := history.Log{Entries: entries}
	data, _ := json.MarshalIndent(log, "", "  ")
	_ = os.WriteFile(path, data, 0644)
	return path
}

func makeArchiveEntry(id, env string) history.Entry {
	return history.Entry{
		ID:           id,
		Environment:  env,
		CapturedAt:   time.Now().UTC(),
		SnapshotPath: "/tmp/snap-" + id + ".json",
		Changes:      []diff.Change{{Key: "x", Old: "1", New: "2"}},
	}
}

func TestRunArchive_NoHistory(t *testing.T) {
	dir := t.TempDir()
	hPath := filepath.Join(dir, "empty.json")
	err := runArchive([]string{"-history", hPath, "-out", filepath.Join(dir, "archive.json")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunArchive_MissingHistoryFile(t *testing.T) {
	dir := t.TempDir()
	err := runArchive([]string{
		"-history", filepath.Join(dir, "nonexistent.json"),
		"-out", filepath.Join(dir, "archive.json"),
	})
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
}

func TestRunArchive_WritesFile(t *testing.T) {
	dir := t.TempDir()
	entries := []history.Entry{
		makeArchiveEntry("abc", "prod"),
		makeArchiveEntry("def", "staging"),
	}
	hPath := writeArchiveHistory(t, dir, entries)
	aPath := filepath.Join(dir, "archive.json")
	if err := runArchive([]string{"-history", hPath, "-out", aPath}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(aPath); err != nil {
		t.Fatalf("archive file not created: %v", err)
	}
}

func TestRunArchive_ListOnly_NoWrite(t *testing.T) {
	dir := t.TempDir()
	entries := []history.Entry{makeArchiveEntry("xyz", "prod")}
	hPath := writeArchiveHistory(t, dir, entries)
	aPath := filepath.Join(dir, "archive.json")
	if err := runArchive([]string{"-history", hPath, "-out", aPath, "-list"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(aPath); err == nil {
		t.Errorf("archive file should not be written in list-only mode")
	}
}

func TestRunArchive_DeduplicatesEntries(t *testing.T) {
	dir := t.TempDir()
	entries := []history.Entry{makeArchiveEntry("dup", "prod")}
	hPath := writeArchiveHistory(t, dir, entries)
	aPath := filepath.Join(dir, "archive.json")
	// Run twice — should not duplicate.
	_ = runArchive([]string{"-history", hPath, "-out", aPath})
	_ = runArchive([]string{"-history", hPath, "-out", aPath})
	data, _ := os.ReadFile(aPath)
	count := strings.Count(string(data), "\"dup\"")
	if count != 1 {
		t.Errorf("expected 1 occurrence of 'dup', got %d", count)
	}
}
