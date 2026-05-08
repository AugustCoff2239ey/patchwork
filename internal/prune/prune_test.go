package prune_test

import (
	"testing"
	"time"

	"github.com/user/patchwork/internal/history"
	"github.com/user/patchwork/internal/prune"
)

func makeLog(envs []string, ages []time.Duration) history.Log {
	log := make(history.Log, len(envs))
	for i, env := range envs {
		log[i] = history.Entry{
			Environment: env,
			Timestamp:   time.Now().Add(-ages[i]),
			SnapshotID:  fmt.Sprintf("snap-%d", i),
		}
	}
	return log
}

import "fmt"

func TestApply_RemovesOldEntries(t *testing.T) {
	log := makeLog(
		[]string{"prod", "prod", "prod"},
		[]time.Duration{1 * time.Hour, 48 * time.Hour, 72 * time.Hour},
	)

	opts := prune.Options{OlderThan: 24 * time.Hour, KeepLast: 0}
	kept, res, err := prune.Apply(log, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 2 {
		t.Errorf("expected 2 removed, got %d", res.Removed)
	}
	if len(kept) != 1 {
		t.Errorf("expected 1 kept entry, got %d", len(kept))
	}
}

func TestApply_RespectsKeepLast(t *testing.T) {
	log := makeLog(
		[]string{"staging", "staging", "staging"},
		[]time.Duration{1 * time.Hour, 48 * time.Hour, 96 * time.Hour},
	)

	opts := prune.Options{OlderThan: 24 * time.Hour, KeepLast: 2}
	kept, res, err := prune.Apply(log, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", res.Removed)
	}
	if len(kept) != 2 {
		t.Errorf("expected 2 kept, got %d", len(kept))
	}
}

func TestApply_FiltersByEnvironment(t *testing.T) {
	log := makeLog(
		[]string{"prod", "dev", "prod"},
		[]time.Duration{96 * time.Hour, 96 * time.Hour, 96 * time.Hour},
	)

	opts := prune.Options{OlderThan: 24 * time.Hour, KeepLast: 0, Environment: "prod"}
	kept, res, err := prune.Apply(log, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 2 {
		t.Errorf("expected 2 removed (prod only), got %d", res.Removed)
	}
	for _, e := range kept {
		if e.Environment != "dev" {
			t.Errorf("expected only dev entries kept, got %s", e.Environment)
		}
	}
}

func TestApply_InvalidKeepLast(t *testing.T) {
	log := history.Log{}
	_, _, err := prune.Apply(log, prune.Options{KeepLast: -1})
	if err == nil {
		t.Error("expected error for negative KeepLast")
	}
}
