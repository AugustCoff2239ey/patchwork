package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.env")
	if err != nil {
		t.Fatalf("create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestRunBaseline_CreatesBaselineFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	configFile := writeTempConfig(t, "db.host=localhost\ndb.port=5432\n")

	if err := runBaseline("test-env", configFile); err != nil {
		t.Fatalf("runBaseline failed: %v", err)
	}

	expectedPath := filepath.Join(dir, ".patchwork", "baselines", "test-env.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected baseline file at %s", expectedPath)
	}
}

func TestRunBaselineDiff_NoDrift(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	configFile := writeTempConfig(t, "db.host=localhost\ndb.port=5432\n")

	if err := runBaseline("staging", configFile); err != nil {
		t.Fatalf("runBaseline failed: %v", err)
	}

	// Same config — no drift expected
	if err := runBaselineDiff("staging", configFile); err != nil {
		t.Fatalf("runBaselineDiff failed: %v", err)
	}
}

func TestRunBaselineDiff_MissingBaseline(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	configFile := writeTempConfig(t, "db.host=localhost\n")

	err := runBaselineDiff("nonexistent-env", configFile)
	if err == nil {
		t.Error("expected error when baseline is missing, got nil")
	}
}
