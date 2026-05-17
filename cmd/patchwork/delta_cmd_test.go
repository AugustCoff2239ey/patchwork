package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/patchwork/internal/snapshot"
)

func writeDeltaSnap(t *testing.T, dir, name string, env string, data map[string]string) string {
	t.Helper()
	snap := snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
	p := filepath.Join(dir, name)
	b, _ := json.Marshal(snap)
	os.WriteFile(p, b, 0644)
	return p
}

func TestRunDelta_MissingBeforeFlag(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected exit on missing --before flag")
		}
	}()
	runDelta([]string{"--after", "some.json"})
}

func TestRunDelta_MissingAfterFlag(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected exit on missing --after flag")
		}
	}()
	runDelta([]string{"--before", "some.json"})
}

func TestRunDelta_EnvironmentMismatch(t *testing.T) {
	dir := t.TempDir()
	before := writeDeltaSnap(t, dir, "before.json", "prod", map[string]string{"x": "1"})
	after := writeDeltaSnap(t, dir, "after.json", "staging", map[string]string{"x": "2"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected exit on environment mismatch")
		}
	}()
	runDelta([]string{"--before", before, "--after", after})
}

func TestRunDelta_NoChanges(t *testing.T) {
	dir := t.TempDir()
	data := map[string]string{"key": "value"}
	before := writeDeltaSnap(t, dir, "before.json", "prod", data)
	after := writeDeltaSnap(t, dir, "after.json", "prod", data)

	// Should not panic or exit
	runDelta([]string{"--before", before, "--after", after})
}

func TestRunDelta_WithChanges(t *testing.T) {
	dir := t.TempDir()
	before := writeDeltaSnap(t, dir, "before.json", "prod", map[string]string{"a": "old", "b": "gone"})
	after := writeDeltaSnap(t, dir, "after.json", "prod", map[string]string{"a": "new", "c": "added"})

	// Should not panic
	runDelta([]string{"--before", before, "--after", after})
}
