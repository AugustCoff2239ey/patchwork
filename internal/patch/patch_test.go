package patch

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
}

func TestApply_SetAddsKey(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"HOST": "localhost"})
	plan := Plan{
		Environment: "prod",
		Ops:         []Op{{Action: "set", Key: "PORT", Value: "8080"}},
	}
	r, err := Apply(plan, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snap.Data["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", r.Snap.Data["PORT"])
	}
	if len(r.Applied) != 1 {
		t.Errorf("expected 1 applied op, got %d", len(r.Applied))
	}
}

func TestApply_DeleteRemovesKey(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"DEBUG": "true", "HOST": "localhost"})
	plan := Plan{
		Environment: "staging",
		Ops:         []Op{{Action: "delete", Key: "DEBUG"}},
	}
	r, err := Apply(plan, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Snap.Data["DEBUG"]; ok {
		t.Error("expected DEBUG to be deleted")
	}
}

func TestApply_DeleteMissingKeyIsSkipped(t *testing.T) {
	snap := makeSnap("dev", map[string]string{"HOST": "localhost"})
	plan := Plan{
		Environment: "dev",
		Ops:         []Op{{Action: "delete", Key: "MISSING"}},
	}
	r, err := Apply(plan, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Skipped) != 1 {
		t.Errorf("expected 1 skipped op, got %d", len(r.Skipped))
	}
}

func TestApply_RenameMovesKey(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"OLD_KEY": "value"})
	plan := Plan{
		Environment: "prod",
		Ops:         []Op{{Action: "rename", Key: "OLD_KEY", NewKey: "NEW_KEY"}},
	}
	r, err := Apply(plan, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snap.Data["NEW_KEY"] != "value" {
		t.Errorf("expected NEW_KEY=value, got %q", r.Snap.Data["NEW_KEY"])
	}
	if _, ok := r.Snap.Data["OLD_KEY"]; ok {
		t.Error("expected OLD_KEY to be removed after rename")
	}
}

func TestApply_EmptyEnvironmentReturnsError(t *testing.T) {
	snap := makeSnap("prod", map[string]string{})
	plan := Plan{Environment: "", Ops: []Op{}}
	_, err := Apply(plan, snap)
	if err == nil {
		t.Error("expected error for empty environment")
	}
}

func TestRender_ContainsAppliedAndSkipped(t *testing.T) {
	r := Result{
		Applied: []string{"set HOST"},
		Skipped: []string{"delete MISSING (not found)"},
		Snap:    makeSnap("prod", map[string]string{}),
	}
	out := Render(r)
	if !strings.Contains(out, "Applied") {
		t.Error("expected output to contain 'Applied'")
	}
	if !strings.Contains(out, "Skipped") {
		t.Error("expected output to contain 'Skipped'")
	}
}
