package merge

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{Environment: env, Timestamp: time.Now(), Data: data}
}

func TestApply_MergesDisjointKeys(t *testing.T) {
	base := makeSnap("prod", map[string]string{"a": "1", "b": "2"})
	incoming := makeSnap("prod", map[string]string{"c": "3", "d": "4"})
	r, err := Apply(base, incoming, StrategyOurs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.MergedKeys != 4 {
		t.Errorf("expected 4 merged keys, got %d", r.MergedKeys)
	}
	if len(r.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(r.Conflicts))
	}
}

func TestApply_StrategyOurs_KeepsBaseValue(t *testing.T) {
	base := makeSnap("prod", map[string]string{"key": "base-val"})
	incoming := makeSnap("prod", map[string]string{"key": "new-val"})
	r, err := Apply(base, incoming, StrategyOurs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snapshot.Data["key"] != "base-val" {
		t.Errorf("expected base-val, got %q", r.Snapshot.Data["key"])
	}
	if len(r.Conflicts) != 1 || r.Conflicts[0].Resolved != "base-val" {
		t.Errorf("expected 1 conflict resolved to base-val")
	}
}

func TestApply_StrategyTheirs_KeepsIncomingValue(t *testing.T) {
	base := makeSnap("prod", map[string]string{"key": "base-val"})
	incoming := makeSnap("prod", map[string]string{"key": "new-val"})
	r, err := Apply(base, incoming, StrategyTheirs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snapshot.Data["key"] != "new-val" {
		t.Errorf("expected new-val, got %q", r.Snapshot.Data["key"])
	}
}

func TestApply_StrategyError_ReturnsError(t *testing.T) {
	base := makeSnap("prod", map[string]string{"key": "v1"})
	incoming := makeSnap("prod", map[string]string{"key": "v2"})
	_, err := Apply(base, incoming, StrategyError)
	if err == nil {
		t.Fatal("expected error on conflict, got nil")
	}
	if !strings.Contains(err.Error(), "merge conflict") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestApply_NoConflictOnIdenticalValues(t *testing.T) {
	base := makeSnap("staging", map[string]string{"x": "same"})
	incoming := makeSnap("staging", map[string]string{"x": "same"})
	r, err := Apply(base, incoming, StrategyError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Conflicts) != 0 {
		t.Errorf("expected no conflicts for identical values")
	}
}

func TestApply_PreservesBaseEnvironment(t *testing.T) {
	base := makeSnap("prod", map[string]string{})
	incoming := makeSnap("staging", map[string]string{})
	r, err := Apply(base, incoming, StrategyOurs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snapshot.Environment != "prod" {
		t.Errorf("expected environment prod, got %q", r.Snapshot.Environment)
	}
}

func TestRender_ContainsConflictInfo(t *testing.T) {
	base := makeSnap("prod", map[string]string{"db_host": "old"})
	incoming := makeSnap("prod", map[string]string{"db_host": "new"})
	r, _ := Apply(base, incoming, StrategyOurs)
	out := Render(r)
	if !strings.Contains(out, "conflict") {
		t.Errorf("expected render to mention conflict, got: %s", out)
	}
	if !strings.Contains(out, "db_host") {
		t.Errorf("expected render to include key name")
	}
}
