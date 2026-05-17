package truncate_test

import (
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/truncate"
)

func makeLog(envs []string, times []time.Time) history.Log {
	var entries []history.Entry
	for i, env := range envs {
		entries = append(entries, history.Entry{
			Environment:  env,
			Timestamp:    times[i],
			SnapshotPath: fmt.Sprintf("/snap/%d.json", i),
		})
	}
	return history.Log{Entries: entries}
}

func TestApply_NoOptionsReturnsError(t *testing.T) {
	log := history.Log{Entries: []history.Entry{}}
	_, _, err := truncate.Apply(log, truncate.Options{})
	if err == nil {
		t.Fatal("expected error when no options provided")
	}
}

func TestApply_RemovesEntriesBeforeTime(t *testing.T) {
	now := time.Now()
	log := makeLog(
		[]string{"prod", "prod", "prod"},
		[]time.Time{now.Add(-3 * time.Hour), now.Add(-1 * time.Hour), now},
	)
	cutoff := now.Add(-2 * time.Hour)
	out, res, err := truncate.Apply(log, truncate.Options{Before: cutoff})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", res.Removed)
	}
	if len(out.Entries) != 2 {
		t.Errorf("expected 2 remaining entries, got %d", len(out.Entries))
	}
}

func TestApply_RespectsMaxEntries(t *testing.T) {
	now := time.Now()
	log := makeLog(
		[]string{"staging", "staging", "staging", "staging"},
		[]time.Time{now.Add(-4 * time.Hour), now.Add(-3 * time.Hour), now.Add(-2 * time.Hour), now.Add(-1 * time.Hour)},
	)
	out, res, err := truncate.Apply(log, truncate.Options{MaxEntries: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 2 {
		t.Errorf("expected 2 removed, got %d", res.Removed)
	}
	if len(out.Entries) != 2 {
		t.Errorf("expected 2 remaining, got %d", len(out.Entries))
	}
}

func TestApply_FiltersByEnvironment(t *testing.T) {
	now := time.Now()
	log := makeLog(
		[]string{"prod", "dev", "prod"},
		[]time.Time{now.Add(-5 * time.Hour), now.Add(-4 * time.Hour), now},
	)
	out, res, err := truncate.Apply(log, truncate.Options{
		Environment: "prod",
		Before:      now.Add(-1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("expected 1 prod entry removed, got %d", res.Removed)
	}
	// dev entry must be preserved
	devCount := 0
	for _, e := range out.Entries {
		if e.Environment == "dev" {
			devCount++
		}
	}
	if devCount != 1 {
		t.Errorf("expected dev entry to be preserved, got %d dev entries", devCount)
	}
}

func TestRender_ContainsSummary(t *testing.T) {
	res := truncate.Result{Removed: 3, Remaining: 7, Environment: "prod"}
	out := truncate.Render(res)
	if out == "" {
		t.Fatal("expected non-empty render output")
	}
	if !contains(out, "prod") {
		t.Errorf("expected environment name in output, got: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
