package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/report"
)

func makeEntry(file string, changes []diff.Change) report.ReportEntry {
	return report.ReportEntry{
		Log: history.LogEntry{
			File:      file,
			Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		Changes: changes,
	}
}

func TestBuild_CountsChanges(t *testing.T) {
	entries := []report.ReportEntry{
		makeEntry("a.json", []diff.Change{
			{Type: diff.Added, Key: "x"},
			{Type: diff.Modified, Key: "y"},
		}),
		makeEntry("b.json", []diff.Change{
			{Type: diff.Removed, Key: "z"},
		}),
	}
	s := report.Build(entries)
	if s.TotalSnapshots != 2 {
		t.Errorf("expected 2 snapshots, got %d", s.TotalSnapshots)
	}
	if s.TotalChanges != 3 {
		t.Errorf("expected 3 total changes, got %d", s.TotalChanges)
	}
	if s.AddedKeys != 1 || s.RemovedKeys != 1 || s.ModifiedKeys != 1 {
		t.Errorf("unexpected breakdown: +%d -%d ~%d", s.AddedKeys, s.RemovedKeys, s.ModifiedKeys)
	}
}

func TestBuild_Empty(t *testing.T) {
	s := report.Build(nil)
	if s.TotalSnapshots != 0 || s.TotalChanges != 0 {
		t.Error("expected zero counts for empty input")
	}
}

func TestRender_ContainsExpectedSections(t *testing.T) {
	entries := []report.ReportEntry{
		makeEntry("cfg.yaml", []diff.Change{
			{Type: diff.Added, Key: "port"},
		}),
	}
	s := report.Build(entries)
	var buf bytes.Buffer
	if err := report.Render(&buf, s); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Patchwork Drift Report", "cfg.yaml", "1 change", "Added keys       : 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nGot:\n%s", want, out)
		}
	}
}
