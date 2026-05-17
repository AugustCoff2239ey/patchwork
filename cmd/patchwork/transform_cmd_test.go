package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork/internal/transform"
)

func writeTransformConfig(t *testing.T, dir string, data map[string]string) string {
	t.Helper()
	b, _ := json.Marshal(data)
	p := filepath.Join(dir, "config.json")
	os.WriteFile(p, b, 0o644)
	return p
}

func writeTransformOps(t *testing.T, dir string, ops []transform.Op) string {
	t.Helper()
	b, _ := json.Marshal(ops)
	p := filepath.Join(dir, "ops.json")
	os.WriteFile(p, b, 0o644)
	return p
}

func TestRunTransform_MissingConfigFlag(t *testing.T) {
	err := runTransform([]string{"--env", "prod", "--ops", "ops.json"})
	if err == nil || err.Error() != "transform: --config is required" {
		t.Errorf("expected config required error, got %v", err)
	}
}

func TestRunTransform_MissingEnvFlag(t *testing.T) {
	err := runTransform([]string{"--config", "cfg.json", "--ops", "ops.json"})
	if err == nil || err.Error() != "transform: --env is required" {
		t.Errorf("expected env required error, got %v", err)
	}
}

func TestRunTransform_MissingOpsFlag(t *testing.T) {
	err := runTransform([]string{"--config", "cfg.json", "--env", "prod"})
	if err == nil || err.Error() != "transform: --ops is required" {
		t.Errorf("expected ops required error, got %v", err)
	}
}

func TestRunTransform_AppliesOpsAndWritesOutput(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTransformConfig(t, dir, map[string]string{"host": "localhost", "debug": "true"})
	ops := writeTransformOps(t, dir, []transform.Op{
		{Kind: transform.OpSet, Key: "host", Value: "prod.example.com"},
		{Kind: transform.OpDelete, Key: "debug"},
	})
	out := filepath.Join(dir, "out.json")
	err := runTransform([]string{"--config", cfg, "--env", "production", "--ops", ops, "--out", out})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestRunTransform_InvalidOpsFile(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTransformConfig(t, dir, map[string]string{"k": "v"})
	badOps := filepath.Join(dir, "bad.json")
	os.WriteFile(badOps, []byte("not-json"), 0o644)
	err := runTransform([]string{"--config", cfg, "--env", "test", "--ops", badOps})
	if err == nil {
		t.Error("expected error for invalid ops JSON")
	}
}
