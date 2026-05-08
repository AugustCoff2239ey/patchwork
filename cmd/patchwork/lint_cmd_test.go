package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeLintConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.env")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write lint config: %v", err)
	}
	return p
}

func TestRunLint_PassesCleanConfig(t *testing.T) {
	path := writeLintConfig(t, "HOST=localhost\nPORT=8080\nREGION=us-east-1\n")
	if err := runLint(path, "prod", false); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRunLint_MissingConfigPath(t *testing.T) {
	err := runLint("", "prod", false)
	if err == nil {
		t.Error("expected error for missing config path")
	}
}

func TestRunLint_MissingEnvironment(t *testing.T) {
	path := writeLintConfig(t, "HOST=localhost\n")
	err := runLint(path, "", false)
	if err == nil {
		t.Error("expected error for missing environment")
	}
}

func TestRunLint_StrictModeFailsOnEmptyValue(t *testing.T) {
	path := writeLintConfig(t, "HOST=localhost\nPORT=\n")
	err := runLint(path, "staging", true)
	if err == nil {
		t.Error("expected strict mode to return error on empty value")
	}
}

func TestRunLint_NonStrictModePassesOnEmptyValue(t *testing.T) {
	path := writeLintConfig(t, "HOST=localhost\nPORT=\n")
	err := runLint(path, "staging", false)
	if err != nil {
		t.Errorf("expected no error in non-strict mode, got: %v", err)
	}
}

func TestRunLint_MissingFile(t *testing.T) {
	err := runLint("/nonexistent/path/config.env", "prod", false)
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
