package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/patchwork/internal/patch"
)

func writePatchConfig(t *testing.T, data map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.env")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	defer f.Close()
	for k, v := range data {
		fmt.Fprintf(f, "%s=%s\n", k, v)
	}
	return path
}

func writePatchOps(t *testing.T, ops []patch.Op) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ops.json")
	b, err := json.Marshal(ops)
	if err != nil {
		t.Fatalf("marshal ops: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write ops: %v", err)
	}
	return path
}

func TestRunPatch_MissingConfigFlag(t *testing.T) {
	err := runPatch([]string{"--env", "prod", "--ops", "ops.json"})
	if err == nil {
		t.Error("expected error for missing --config")
	}
}

func TestRunPatch_MissingEnvFlag(t *testing.T) {
	err := runPatch([]string{"--config", "cfg.env", "--ops", "ops.json"})
	if err == nil {
		t.Error("expected error for missing --env")
	}
}

func TestRunPatch_MissingOpsFlag(t *testing.T) {
	err := runPatch([]string{"--config", "cfg.env", "--env", "prod"})
	if err == nil {
		t.Error("expected error for missing --ops")
	}
}

func TestRunPatch_DryRunDoesNotWriteFile(t *testing.T) {
	cfg := writePatchConfig(t, map[string]string{"HOST": "localhost"})
	ops := writePatchOps(t, []patch.Op{
		{Action: "set", Key: "PORT", Value: "9090"},
	})
	err := runPatch([]string{
		"--config", cfg,
		"--env", "prod",
		"--ops", ops,
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(cfg + ".patched"); !os.IsNotExist(statErr) {
		t.Error("expected no patched file to be written in dry-run mode")
	}
}

func TestRunPatch_WritesOutputFile(t *testing.T) {
	cfg := writePatchConfig(t, map[string]string{"HOST": "localhost"})
	ops := writePatchOps(t, []patch.Op{
		{Action: "set", Key: "PORT", Value: "9090"},
	})
	err := runPatch([]string{
		"--config", cfg,
		"--env", "prod",
		"--ops", ops,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(cfg + ".patched"); os.IsNotExist(statErr) {
		t.Error("expected patched file to be written")
	}
}
