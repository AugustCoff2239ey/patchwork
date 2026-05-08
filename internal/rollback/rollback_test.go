package rollback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

func makeEntry(env, snapPath string, t time.Time) history.Entry {
	return history.Entry{
		Environment:  env,
		SnapshotPath: snapPath,
		CapturedAt:   t,
	}
}

func writeSnapFile(t *testing.T, dir string, data map[string]string) string {
	t.Helper()
	snap := snapshot.Snapshot{Data: data}
	path := filepath.Join(dir, "snap_"+time.Now().Format("150405.000000000")+".json")
	b, _ := json.Marshal(snap)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("writeSnapFile: %v", err)
	}
	return path
}

func TestPrepare_ReturnsPlan(t *testing.T) {
	now := time.Now().UTC()
	log := history.Log{
		Entries: []history.Entry{
			makeEntry("prod", "/old.json", now.Add(-2*time.Hour)),
			makeEntry("prod", "/new.json", now.Add(-1*time.Hour)),
		},
	}
	plan, err := Prepare("prod", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.FromEntry.SnapshotPath != "/new.json" {
		t.Errorf("expected FromEntry /new.json, got %s", plan.FromEntry.SnapshotPath)
	}
	if plan.ToEntry.SnapshotPath != "/old.json" {
		t.Errorf("expected ToEntry /old.json, got %s", plan.ToEntry.SnapshotPath)
	}
}

func TestPrepare_InsufficientEntries(t *testing.T) {
	log := history.Log{
		Entries: []history.Entry{
			makeEntry("staging", "/snap.json", time.Now()),
		},
	}
	_, err := Prepare("staging", log)
	if err == nil {
		t.Fatal("expected error for insufficient entries")
	}
}

func TestPrepare_FiltersByEnvironment(t *testing.T) {
	now := time.Now().UTC()
	log := history.Log{
		Entries: []history.Entry{
			makeEntry("prod", "/p1.json", now.Add(-3*time.Hour)),
			makeEntry("dev", "/d1.json", now.Add(-2*time.Hour)),
			makeEntry("prod", "/p2.json", now.Add(-1*time.Hour)),
		},
	}
	plan, err := Prepare("prod", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Environment != "prod" {
		t.Errorf("expected environment prod, got %s", plan.Environment)
	}
}

func TestApply_WritesSnapshot(t *testing.T) {
	dir := t.TempDir()
	srcPath := writeSnapFile(t, dir, map[string]string{"key": "value"})
	now := time.Now().UTC()
	plan := Plan{
		Environment: "prod",
		FromEntry:   makeEntry("prod", "/current.json", now),
		ToEntry:     makeEntry("prod", srcPath, now.Add(-1*time.Hour)),
		CreatedAt:   now,
	}
	destPath := filepath.Join(dir, "rollback_result.json")
	result, err := Apply(plan, destPath)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if !result.Applied {
		t.Error("expected Applied=true")
	}
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("expected destination file to exist")
	}
}
