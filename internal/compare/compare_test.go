package compare

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeCompareSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
}

func TestEnvironments_ReturnsSortedNames(t *testing.T) {
	snaps := []snapshot.Snapshot{
		makeCompareSnap("production", map[string]string{}),
		makeCompareSnap("staging", map[string]string{}),
		makeCompareSnap("dev", map[string]string{}),
	}
	envs := Environments(snaps)
	if len(envs) != 3 {
		t.Fatalf("expected 3 environments, got %d", len(envs))
	}
	if envs[0] != "dev" || envs[1] != "production" || envs[2] != "staging" {
		t.Errorf("unexpected order: %v", envs)
	}
}

func TestRender_ContainsEnvironmentHeaders(t *testing.T) {
	snaps := []snapshot.Snapshot{
		makeCompareSnap("staging", map[string]string{"key": "val"}),
		makeCompareSnap("production", map[string]string{"key": "other"}),
	}
	out := Render(snaps)
	if !strings.Contains(out, "staging") {
		t.Errorf("render missing staging: %s", out)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("render missing production: %s", out)
	}
}

func TestRender_EmptySnapshots(t *testing.T) {
	out := Render([]snapshot.Snapshot{})
	if out == "" {
		t.Error("expected non-empty render output even for empty input")
	}
}

func TestRender_SharedKeyShowsDifference(t *testing.T) {
	snaps := []snapshot.Snapshot{
		makeCompareSnap("a", map[string]string{"shared": "x"}),
		makeCompareSnap("b", map[string]string{"shared": "y"}),
	}
	out := Render(snaps)
	if !strings.Contains(out, "shared") {
		t.Errorf("expected shared key in output: %s", out)
	}
}
