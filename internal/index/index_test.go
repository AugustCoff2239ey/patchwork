package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/index"
	"github.com/patchwork/internal/snapshot"
)

func makeEntry(id, env string, data map[string]string) history.Entry {
	return history.Entry{
		ID: id,
		Snapshot: snapshot.Snapshot{
			Environment: env,
			Timestamp:   "2024-01-01T00:00:00Z",
			Data:        data,
		},
	}
}

func TestBuild_IndexesAllKeys(t *testing.T) {
	log := []history.Entry{
		makeEntry("a1", "prod", map[string]string{"host": "example.com", "port": "8080"}),
		makeEntry("b1", "staging", map[string]string{"host": "staging.example.com"}),
	}
	idx := index.Build(log)
	if len(idx.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(idx.Entries))
	}
}

func TestBuild_SortedByEnvironmentAndKey(t *testing.T) {
	log := []history.Entry{
		makeEntry("z1", "staging", map[string]string{"z": "1"}),
		makeEntry("a1", "prod", map[string]string{"a": "1"}),
	}
	idx := index.Build(log)
	if idx.Entries[0].Environment != "prod" {
		t.Errorf("expected prod first, got %s", idx.Entries[0].Environment)
	}
}

func TestSearch_MatchesByKey(t *testing.T) {
	log := []history.Entry{
		makeEntry("a1", "prod", map[string]string{"database_url": "postgres://", "port": "5432"}),
	}
	idx := index.Build(log)
	results := index.Search(idx, "database", "")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "database_url" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestSearch_FiltersByEnvironment(t *testing.T) {
	log := []history.Entry{
		makeEntry("a1", "prod", map[string]string{"host": "prod.example.com"}),
		makeEntry("b1", "staging", map[string]string{"host": "staging.example.com"}),
	}
	idx := index.Build(log)
	results := index.Search(idx, "host", "prod")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Environment != "prod" {
		t.Errorf("unexpected env: %s", results[0].Environment)
	}
}

func TestSearch_NoMatch_ReturnsEmpty(t *testing.T) {
	log := []history.Entry{
		makeEntry("a1", "prod", map[string]string{"key": "value"}),
	}
	idx := index.Build(log)
	results := index.Search(idx, "nonexistent", "")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	log := []history.Entry{
		makeEntry("a1", "prod", map[string]string{"key": "val"}),
	}
	idx := index.Build(log)
	path := filepath.Join(t.TempDir(), "index.json")
	if err := index.Save(idx, path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := index.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Entries) != len(idx.Entries) {
		t.Errorf("entry count mismatch: got %d, want %d", len(loaded.Entries), len(idx.Entries))
	}
}

func TestLoad_MissingFile_ReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	idx, err := index.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("expected empty index, got %d entries", len(idx.Entries))
	}
}

func TestRender_NoResults(t *testing.T) {
	out := index.Render(nil)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}

func TestRender_ContainsKeyAndEnv(t *testing.T) {
	entries := []index.Entry{
		{Environment: "prod", Key: "host", Value: "example.com", SnapshotID: "abc", Timestamp: "2024-01-01T00:00:00Z"},
	}
	out := index.Render(entries)
	for _, want := range []string{"prod", "host", "example.com"} {
		if !containsStr(out, want) {
			t.Errorf("render output missing %q", want)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func init() {
	_ = os.Getenv // suppress unused import
}
