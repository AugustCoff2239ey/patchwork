package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/patchwork/internal/audit"
)

func writeAuditLog(t *testing.T, dir string, events []audit.Event) string {
	t.Helper()
	path := filepath.Join(dir, "audit.json")
	log := &audit.Log{Events: events}
	data, _ := json.MarshalIndent(log, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writeAuditLog: %v", err)
	}
	return path
}

func TestRunAudit_MissingFile(t *testing.T) {
	// Should print "No audit events found." when file is absent.
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	// Capture stdout by redirecting — here we just call recordAudit and verify no panic.
	recordAudit(path, audit.EventCapture, "dev", "test capture")

	loaded, err := audit.Load(path)
	if err != nil {
		t.Fatalf("Load after record: %v", err)
	}
	if len(loaded.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(loaded.Events))
	}
}

func TestRecordAudit_AppendsEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	recordAudit(path, audit.EventDiff, "staging", "ran diff")
	recordAudit(path, audit.EventExport, "staging", "exported")

	log, err := audit.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(log.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(log.Events))
	}
}

func TestRecordAudit_FilterByEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	recordAudit(path, audit.EventCapture, "production", "prod cap")
	recordAudit(path, audit.EventCapture, "staging", "stage cap")

	log, _ := audit.Load(path)
	results := audit.Filter(log, "", "production")
	if len(results) != 1 {
		t.Errorf("expected 1 production event, got %d", len(results))
	}
}

func TestRenderAudit_ContainsMessage(t *testing.T) {
	events := []audit.Event{
		{Kind: audit.EventRollback, Environment: "prod", Message: "rolled back to v3"},
	}
	out := audit.Render(events)
	if !strings.Contains(out, "rolled back to v3") {
		t.Errorf("expected message in output, got: %q", out)
	}
	if !strings.Contains(out, "rollback") {
		t.Errorf("expected kind in output, got: %q", out)
	}
}
