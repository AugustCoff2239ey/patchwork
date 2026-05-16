package promote

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

func TestApply_PromotesAllKeys(t *testing.T) {
	src := makeSnap("staging", map[string]string{"a": "1", "b": "2"})
	dst := makeSnap("production", map[string]string{})

	result, plan, err := Apply(src, dst, Options{Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Applied != 2 {
		t.Errorf("expected 2 applied, got %d", plan.Applied)
	}
	if result.Data["a"] != "1" || result.Data["b"] != "2" {
		t.Errorf("unexpected result data: %v", result.Data)
	}
}

func TestApply_SkipsExistingWithoutOverwrite(t *testing.T) {
	src := makeSnap("staging", map[string]string{"a": "new", "b": "2"})
	dst := makeSnap("production", map[string]string{"a": "old"})

	_, plan, err := Apply(src, dst, Options{Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", plan.Skipped)
	}
	if plan.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", plan.Applied)
	}
}

func TestApply_OnlyKeys_FiltersKeys(t *testing.T) {
	src := makeSnap("staging", map[string]string{"a": "1", "b": "2", "c": "3"})
	dst := makeSnap("production", map[string]string{})

	_, plan, err := Apply(src, dst, Options{OnlyKeys: []string{"a", "c"}, Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Applied != 2 {
		t.Errorf("expected 2 applied, got %d", plan.Applied)
	}
}

func TestApply_EmptySourceEnv_ReturnsError(t *testing.T) {
	src := makeSnap("", map[string]string{"a": "1"})
	dst := makeSnap("production", map[string]string{})

	_, _, err := Apply(src, dst, Options{})
	if err == nil {
		t.Fatal("expected error for empty source environment")
	}
}

func TestApply_SameEnvWithoutOverwrite_ReturnsError(t *testing.T) {
	src := makeSnap("prod", map[string]string{"a": "1"})
	dst := makeSnap("prod", map[string]string{})

	_, _, err := Apply(src, dst, Options{Overwrite: false})
	if err == nil {
		t.Fatal("expected error when source and target env are the same")
	}
}

func TestRender_ContainsEnvironments(t *testing.T) {
	p := Plan{SourceEnv: "staging", TargetEnv: "production", Applied: 3, Skipped: 1, Keys: []string{"x", "y", "z"}}
	out := Render(p)
	if !strings.Contains(out, "staging") || !strings.Contains(out, "production") {
		t.Errorf("render missing environment names: %s", out)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("render missing applied count: %s", out)
	}
}
