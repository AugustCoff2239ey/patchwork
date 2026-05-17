package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/patchwork/internal/scope"
)

func writeScopeConfig(t *testing.T, data map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "scope-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func writeScopeFile(t *testing.T, sc scope.Scope) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "scope-def-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(sc); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestRunScope_MissingNameFlag(t *testing.T) {
	cfg := writeScopeConfig(t, map[string]string{"host": "localhost"})
	err := runScope([]string{"--config", cfg, "--env", "prod", "--keys", "host"})
	if err == nil || !strings.Contains(err.Error(), "--name") {
		t.Errorf("expected --name error, got %v", err)
	}
}

func TestRunScope_MissingKeysFlag(t *testing.T) {
	cfg := writeScopeConfig(t, map[string]string{"host": "localhost"})
	err := runScope([]string{"--config", cfg, "--env", "prod", "--name", "net"})
	if err == nil || !strings.Contains(err.Error(), "--keys") {
		t.Errorf("expected --keys error, got %v", err)
	}
}

func TestRunScope_WithScopeFile(t *testing.T) {
	cfg := writeScopeConfig(t, map[string]string{"host": "localhost", "port": "8080"})
	sc := scope.Scope{Name: "net", Environment: "prod", Keys: []string{"host", "port"}}
	sf := writeScopeFile(t, sc)

	err := runScope([]string{"--config", cfg, "--env", "prod", "--scope-file", sf})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunScope_MissingConfigFile(t *testing.T) {
	err := runScope([]string{"--config", "/nonexistent/path.json", "--env", "prod", "--name", "net", "--keys", "host"})
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
