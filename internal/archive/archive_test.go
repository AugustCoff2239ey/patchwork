package archive_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/archive"
	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
)

func makeLog(envs ...string) history.Log {
	log := history.Log{}
	for i, env := range envs {
		log.Entries = append(log.Entries, history.Entry{
			ID:           fmt.Sprintf("id-%d", i),
			Environment:  env,
			CapturedAt:   time.Now().UTC(),
			SnapshotPath: "/tmp/snap-" + env + ".json",
			Changes:      []diff.Change{{Key: "k", Old: "a", New: "b"}},
		})
	}
	return log
}

func TestBuild_IncludesAllEnvs(t *testing.T) {
	log := makeLog("prod", "staging")
	a := archive.Build(log, "")
	if len(a.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(a.Entries))
	}
}

func TestBuild_FiltersByEnvironment(t *testing.T) {
	log := makeLog("prod", "staging", "prod")
	a := archive.Build(log, "prod")
	if len(a.Entries) != 2 {
		t.Fatalf("expected 2 prod entries, got %d", len(a.Entries))
	}
	for _, e := range a.Entries {
		if e.Environment != "prod" {
			t.Errorf("unexpected env: %s", e.Environment)
		}
	}
}

func TestBuild_Empty(t *testing.T) {
	a := archive.Build(history.Log{}, "")
	if len(a.Entries) != 0 {
		t.Fatalf("expected empty archive")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "archive.json")
	log := makeLog("prod")
	a := archive.Build(log, "")
	if err := archive.Save(path, a); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := archive.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Entries) != len(a.Entries) {
		t.Errorf("entry count mismatch: got %d want %d", len(loaded.Entries), len(a.Entries))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	a, err := archive.Load("/nonexistent/archive.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(a.Entries) != 0 {
		t.Errorf("expected empty archive")
	}
}

func TestRender_ContainsEnvironment(t *testing.T) {
	log := makeLog("prod")
	a := archive.Build(log, "")
	out := archive.Render(a)
	if !strings.Contains(out, "prod") {
		t.Errorf("expected 'prod' in render output, got: %s", out)
	}
}

func TestRender_Empty(t *testing.T) {
	out := archive.Render(archive.Archive{})
	if !strings.Contains(out, "no entries") {
		t.Errorf("expected 'no entries' message, got: %s", out)
	}
}
