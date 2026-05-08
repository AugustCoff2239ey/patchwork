package summary

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
)

func makeEntries() []history.Entry {
	return []history.Entry{
		{
			Environment: "production",
			Timestamp:   time.Now(),
			Changes: []diff.Change{
				{Type: "added", Key: "DB_HOST"},
				{Type: "modified", Key: "DB_PORT"},
			},
		},
		{
			Environment: "production",
			Timestamp:   time.Now(),
			Changes: []diff.Change{
				{Type: "removed", Key: "DB_HOST"},
				{Type: "modified", Key: "DB_HOST"},
			},
		},
		{
			Environment: "staging",
			Timestamp:   time.Now(),
			Changes: []diff.Change{
				{Type: "added", Key: "API_KEY"},
			},
		},
	}
}

func TestBuild_CountsByEnvironment(t *testing.T) {
	entries := makeEntries()
	s := Build(entries, "production")

	if s.TotalEntries != 2 {
		t.Errorf("expected 2 entries, got %d", s.TotalEntries)
	}
	if s.TotalAdded != 1 {
		t.Errorf("expected 1 added, got %d", s.TotalAdded)
	}
	if s.TotalRemoved != 1 {
		t.Errorf("expected 1 removed, got %d", s.TotalRemoved)
	}
	if s.TotalModified != 1 {
		t.Errorf("expected 1 modified, got %d", s.TotalModified)
	}
}

func TestBuild_AllEnvironments(t *testing.T) {
	entries := makeEntries()
	s := Build(entries, "")

	if s.TotalEntries != 3 {
		t.Errorf("expected 3 entries, got %d", s.TotalEntries)
	}
	if s.TotalAdded != 2 {
		t.Errorf("expected 2 added, got %d", s.TotalAdded)
	}
}

func TestBuild_MostChangedKey(t *testing.T) {
	entries := makeEntries()
	s := Build(entries, "production")

	if s.MostChanged != "DB_HOST" {
		t.Errorf("expected DB_HOST as most changed, got %s", s.MostChanged)
	}
}

func TestBuild_Empty(t *testing.T) {
	s := Build([]history.Entry{}, "production")

	if s.TotalEntries != 0 {
		t.Errorf("expected 0 entries")
	}
	if s.MostChanged != "" {
		t.Errorf("expected empty MostChanged")
	}
}

func TestRender_ContainsKeyFields(t *testing.T) {
	entries := makeEntries()
	s := Build(entries, "production")
	out := Render(s)

	for _, want := range []string{"production", "Entries", "Added", "Removed", "Modified", "DB_HOST"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}
