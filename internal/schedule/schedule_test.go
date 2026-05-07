package schedule

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsDue_FirstRun(t *testing.T) {
	e := Entry{
		Environment: "prod",
		ConfigPath:  "/etc/app.conf",
		Interval:    5 * time.Minute,
	}
	if !e.IsDue(time.Now()) {
		t.Error("expected entry with zero LastRun to be due")
	}
}

func TestIsDue_NotYetDue(t *testing.T) {
	now := time.Now()
	e := Entry{
		Interval: 10 * time.Minute,
		LastRun:  now.Add(-3 * time.Minute),
	}
	if e.IsDue(now) {
		t.Error("expected entry to not be due yet")
	}
}

func TestIsDue_ExactlyDue(t *testing.T) {
	now := time.Now()
	e := Entry{
		Interval: 5 * time.Minute,
		LastRun:  now.Add(-5 * time.Minute),
	}
	if !e.IsDue(now) {
		t.Error("expected entry to be due at exact interval boundary")
	}
}

func TestMarkRun_UpdatesLastRun(t *testing.T) {
	now := time.Now()
	e := Entry{}
	e.MarkRun(now)
	if !e.LastRun.Equal(now) {
		t.Errorf("expected LastRun %v, got %v", now, e.LastRun)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schedule.json")

	s := &Schedule{
		Entries: []Entry{
			{Environment: "staging", ConfigPath: "/etc/svc.yaml", Interval: 2 * time.Minute},
			{Environment: "prod", ConfigPath: "/etc/app.conf", Interval: 10 * time.Minute},
		},
	}

	if err := Save(path, s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Environment != "staging" {
		t.Errorf("expected staging, got %s", loaded.Entries[0].Environment)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	s, err := Load("/nonexistent/schedule.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(s.Entries) != 0 {
		t.Error("expected empty schedule for missing file")
	}
}

func TestDueEntries_ReturnsOnlyDue(t *testing.T) {
	now := time.Now()
	s := &Schedule{
		Entries: []Entry{
			{Environment: "prod", Interval: 5 * time.Minute, LastRun: now.Add(-10 * time.Minute)},
			{Environment: "dev", Interval: 30 * time.Minute, LastRun: now.Add(-1 * time.Minute)},
			{Environment: "staging", Interval: 1 * time.Minute},
		},
	}

	due := DueEntries(s, now)
	if len(due) != 2 {
		t.Errorf("expected 2 due entries, got %d", len(due))
	}

	_ = os.Getenv // suppress unused import
}
