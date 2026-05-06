package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/patchwork/internal/history"
)

func TestLog_AddAndLatest(t *testing.T) {
	var log history.Log
	log.Add("baseline", "/etc/app.conf", "/snapshots/app_1.json")
	log.Add("after-deploy", "/etc/app.conf", "/snapshots/app_2.json")
	log.Add("baseline", "/etc/db.conf", "/snapshots/db_1.json")

	latest := log.Latest("/etc/app.conf")
	if latest == nil {
		t.Fatal("expected an entry, got nil")
	}
	if latest.Label != "after-deploy" {
		t.Errorf("expected label 'after-deploy', got %q", latest.Label)
	}
}

func TestLog_Latest_NotFound(t *testing.T) {
	var log history.Log
	log.Add("baseline", "/etc/app.conf", "/snapshots/app_1.json")

	if got := log.Latest("/etc/missing.conf"); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestLog_Sorted(t *testing.T) {
	var log history.Log
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)
	t3 := time.Now()

	log.Entries = []history.Entry{
		{Timestamp: t3, Label: "c"},
		{Timestamp: t1, Label: "a"},
		{Timestamp: t2, Label: "b"},
	}

	sorted := log.Sorted()
	if sorted[0].Label != "a" || sorted[1].Label != "b" || sorted[2].Label != "c" {
		t.Errorf("unexpected sort order: %v", sorted)
	}
	// original slice must be unchanged
	if log.Entries[0].Label != "c" {
		t.Error("original entries were mutated")
	}
}

func TestSaveAndLoadLog_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "patchwork", "history.json")

	var log history.Log
	log.Add("v1", "/etc/app.conf", "/snapshots/app_1.json")
	log.Add("v2", "/etc/app.conf", "/snapshots/app_2.json")

	if err := history.SaveLog(path, &log); err != nil {
		t.Fatalf("SaveLog: %v", err)
	}

	loaded, err := history.LoadLog(path)
	if err != nil {
		t.Fatalf("LoadLog: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[1].Label != "v2" {
		t.Errorf("unexpected label: %q", loaded.Entries[1].Label)
	}
}

func TestLoadLog_MissingFile(t *testing.T) {
	log, err := history.LoadLog(filepath.Join(os.TempDir(), "nonexistent_pw_history.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(log.Entries) != 0 {
		t.Errorf("expected empty log, got %d entries", len(log.Entries))
	}
}
