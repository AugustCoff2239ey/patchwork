package export_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/alert"
	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/export"
	"github.com/patchwork/internal/history"
)

func makeEntry(env string) history.Entry {
	return history.Entry{
		Environment: env,
		Timestamp:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func makeChanges() []diff.Change {
	return []diff.Change{
		{Key: "db.host", Type: diff.Added, NewValue: "localhost"},
		{Key: "db.port", Type: diff.Modified, OldValue: "5432", NewValue: "5433"},
		{Key: "cache.ttl", Type: diff.Removed, OldValue: "300"},
	}
}

func TestBuild_SummaryCounts(t *testing.T) {
	rec := export.Build(makeEntry("prod"), makeChanges(), nil)

	if rec.Summary.Added != 1 {
		t.Errorf("expected Added=1, got %d", rec.Summary.Added)
	}
	if rec.Summary.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", rec.Summary.Removed)
	}
	if rec.Summary.Modified != 1 {
		t.Errorf("expected Modified=1, got %d", rec.Summary.Modified)
	}
	if rec.Summary.Total != 3 {
		t.Errorf("expected Total=3, got %d", rec.Summary.Total)
	}
}

func TestBuild_PreservesEnvironment(t *testing.T) {
	rec := export.Build(makeEntry("staging"), nil, nil)
	if rec.Environment != "staging" {
		t.Errorf("expected environment=staging, got %s", rec.Environment)
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	alerts := []alert.Alert{{Severity: alert.Critical, Message: "host changed"}}
	rec := export.Build(makeEntry("prod"), makeChanges(), alerts)

	path, err := export.Write(rec, dir, export.FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}

	var decoded export.ExportRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Summary.Total != 3 {
		t.Errorf("decoded total mismatch: got %d", decoded.Summary.Total)
	}
	if len(decoded.Alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(decoded.Alerts))
	}
}

func TestWrite_TextFormat(t *testing.T) {
	dir := t.TempDir()
	rec := export.Build(makeEntry("dev"), makeChanges(), nil)

	path, err := export.Write(rec, dir, export.FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(filepath.Base(path), ".text") {
		t.Errorf("expected .text extension, got %s", path)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "dev") {
		t.Error("expected environment name in text output")
	}
	if !strings.Contains(string(content), "db.host") {
		t.Error("expected change key in text output")
	}
}

func TestWrite_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	rec := export.Build(makeEntry("prod"), nil, nil)
	_, err := export.Write(rec, dir, export.Format("xml"))
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}
