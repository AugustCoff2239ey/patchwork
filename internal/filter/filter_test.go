package filter_test

import (
	"testing"

	"github.com/patchwork/internal/filter"
	"github.com/patchwork/internal/history"
)

func makeEntries() []history.Entry {
	return []history.Entry{
		{Environment: "production", Timestamp: "2024-03-01T10:00:00Z", ChangeCount: 3},
		{Environment: "staging", Timestamp: "2024-03-05T12:00:00Z", ChangeCount: 0},
		{Environment: "production", Timestamp: "2024-04-01T09:00:00Z", ChangeCount: 1},
		{Environment: "dev", Timestamp: "2024-04-10T08:00:00Z", ChangeCount: 0},
	}
}

func TestApply_FilterByEnvironment(t *testing.T) {
	result := filter.Apply(makeEntries(), filter.Options{Environment: "production"})
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	for _, e := range result {
		if e.Environment != "production" {
			t.Errorf("unexpected environment: %s", e.Environment)
		}
	}
}

func TestApply_FilterBySince(t *testing.T) {
	result := filter.Apply(makeEntries(), filter.Options{Since: "2024-04"})
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestApply_FilterHasChanges(t *testing.T) {
	result := filter.Apply(makeEntries(), filter.Options{HasChanges: true})
	if len(result) != 2 {
		t.Fatalf("expected 2 entries with changes, got %d", len(result))
	}
	for _, e := range result {
		if e.ChangeCount == 0 {
			t.Errorf("entry with zero changes should be excluded")
		}
	}
}

func TestApply_CombinedFilters(t *testing.T) {
	opts := filter.Options{Environment: "production", Since: "2024-04", HasChanges: true}
	result := filter.Apply(makeEntries(), opts)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
}

func TestApply_NoFilters_ReturnsAll(t *testing.T) {
	entries := makeEntries()
	result := filter.Apply(entries, filter.Options{})
	if len(result) != len(entries) {
		t.Fatalf("expected all %d entries, got %d", len(entries), len(result))
	}
}

func TestEnvironments_ReturnsSorted(t *testing.T) {
	envs := filter.Environments(makeEntries())
	expected := []string{"dev", "production", "staging"}
	if len(envs) != len(expected) {
		t.Fatalf("expected %d envs, got %d", len(expected), len(envs))
	}
	for i, e := range expected {
		if envs[i] != e {
			t.Errorf("position %d: expected %s, got %s", i, e, envs[i])
		}
	}
}
