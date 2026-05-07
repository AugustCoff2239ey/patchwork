package baseline_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/patchwork/internal/baseline"
	"github.com/user/patchwork/internal/snapshot"
)

func makeSnapshot() snapshot.Snapshot {
	return snapshot.Snapshot{
		File:      "/etc/app/config.yaml",
		CapturedAt: time.Now().UTC(),
		Values: map[string]string{
			"db.host": "localhost",
			"db.port": "5432",
		},
	}
}

func TestPin_SetsEnvironmentAndSnapshot(t *testing.T) {
	snap := makeSnapshot()
	b := baseline.Pin("production", snap)

	if b.Environment != "production" {
		t.Errorf("expected environment 'production', got %q", b.Environment)
	}
	if b.Snapshot.File != snap.File {
		t.Errorf("expected snapshot file %q, got %q", snap.File, b.Snapshot.File)
	}
	if b.PinnedAt.IsZero() {
		t.Error("expected PinnedAt to be set")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	original := baseline.Pin("staging", makeSnapshot())
	if err := baseline.Save(path, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := baseline.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Environment != original.Environment {
		t.Errorf("environment mismatch: got %q, want %q", loaded.Environment, original.Environment)
	}
	if loaded.Snapshot.File != original.Snapshot.File {
		t.Errorf("snapshot file mismatch: got %q, want %q", loaded.Snapshot.File, original.Snapshot.File)
	}
	if len(loaded.Snapshot.Values) != len(original.Snapshot.Values) {
		t.Errorf("values length mismatch: got %d, want %d", len(loaded.Snapshot.Values), len(original.Snapshot.Values))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := baseline.Load("/nonexistent/baseline.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	b := baseline.Pin("dev", makeSnapshot())
	if err := baseline.Save(path, b); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to exist after Save")
	}
}
