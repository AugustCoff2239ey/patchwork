package clone_test

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/clone"
	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		CapturedAt:  time.Now().UTC(),
		Data:        data,
	}
}

func TestApply_ClonesAllKeys(t *testing.T) {
	src := makeSnap("staging", map[string]string{"DB_HOST": "localhost", "PORT": "5432"})
	res, err := clone.Apply(src, "production", clone.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Snapshot.Environment != "production" {
		t.Errorf("expected environment %q, got %q", "production", res.Snapshot.Environment)
	}
	if len(res.Snapshot.Data) != 2 {
		t.Errorf("expected 2 keys, got %d", len(res.Snapshot.Data))
	}
}

func TestApply_ExcludesKeys(t *testing.T) {
	src := makeSnap("staging", map[string]string{"DB_HOST": "localhost", "SECRET": "s3cr3t", "PORT": "5432"})
	res, err := clone.Apply(src, "production", clone.Options{ExcludeKeys: []string{"SECRET"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.Snapshot.Data["SECRET"]; ok {
		t.Error("expected SECRET to be excluded from clone")
	}
	if len(res.Snapshot.Data) != 2 {
		t.Errorf("expected 2 keys after exclusion, got %d", len(res.Snapshot.Data))
	}
}

func TestApply_EmptyTargetEnv_ReturnsError(t *testing.T) {
	src := makeSnap("staging", map[string]string{"KEY": "val"})
	_, err := clone.Apply(src, "", clone.Options{})
	if err == nil {
		t.Fatal("expected error for empty target environment")
	}
}

func TestApply_SameEnvWithoutOverwrite_ReturnsError(t *testing.T) {
	src := makeSnap("staging", map[string]string{"KEY": "val"})
	_, err := clone.Apply(src, "staging", clone.Options{Overwrite: false})
	if err == nil {
		t.Fatal("expected error when source and target environments are the same")
	}
}

func TestApply_SameEnvWithOverwrite_Succeeds(t *testing.T) {
	src := makeSnap("staging", map[string]string{"KEY": "val"})
	_, err := clone.Apply(src, "staging", clone.Options{Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error with overwrite enabled: %v", err)
	}
}

func TestRender_ContainsEnvironments(t *testing.T) {
	src := makeSnap("staging", map[string]string{"A": "1"})
	res, _ := clone.Apply(src, "production", clone.Options{})
	out := clone.Render(res)
	if !strings.Contains(out, "staging") {
		t.Error("expected render output to contain source environment")
	}
	if !strings.Contains(out, "production") {
		t.Error("expected render output to contain target environment")
	}
}
