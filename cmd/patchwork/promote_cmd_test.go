package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func writePromoteSnap(t *testing.T, dir, name string, env string, data map[string]string) string {
	t.Helper()
	snap := snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(snap); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunPromote_MissingSrcFlag(t *testing.T) {
	err := runPromote([]string{"--env", "production"})
	if err == nil || err.Error() != "--src is required" {
		t.Errorf("expected --src error, got %v", err)
	}
}

func TestRunPromote_MissingEnvFlag(t *testing.T) {
	dir := t.TempDir()
	src := writePromoteSnap(t, dir, "src.json", "staging", map[string]string{"k": "v"})
	err := runPromote([]string{"--src", src})
	if err == nil || err.Error() != "--env is required" {
		t.Errorf("expected --env error, got %v", err)
	}
}

func TestRunPromote_EmptyTarget_WritesOutput(t *testing.T) {
	dir := t.TempDir()
	src := writePromoteSnap(t, dir, "src.json", "staging", map[string]string{"key": "val"})
	out := filepath.Join(dir, "result.json")

	err := runPromote([]string{"--src", src, "--env", "production", "--out", out})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var snap snapshot.Snapshot
	f, _ := os.Open(out)
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if snap.Data["key"] != "val" {
		t.Errorf("expected promoted key, got %v", snap.Data)
	}
	if snap.Environment != "production" {
		t.Errorf("expected production env, got %s", snap.Environment)
	}
}

func TestRunPromote_OnlyKeys_FiltersKeys(t *testing.T) {
	dir := t.TempDir()
	src := writePromoteSnap(t, dir, "src.json", "staging", map[string]string{"a": "1", "b": "2"})
	out := filepath.Join(dir, "result.json")

	err := runPromote([]string{"--src", src, "--env", "production", "--keys", "a", "--out", out})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var snap snapshot.Snapshot
	f, _ := os.Open(out)
	defer f.Close()
	json.NewDecoder(f).Decode(&snap)
	if _, ok := snap.Data["b"]; ok {
		t.Errorf("key 'b' should not have been promoted")
	}
}
