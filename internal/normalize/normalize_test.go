package normalize_test

import (
	"testing"
	"time"

	"github.com/patchwork/internal/normalize"
	"github.com/patchwork/internal/snapshot"
)

func makeSnap(data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: "test",
		Timestamp:   time.Now(),
		Data:        data,
	}
}

func TestApply_TrimSpaceValues(t *testing.T) {
	snap := makeSnap(map[string]string{
		"key": "  value  ",
	})
	opts := normalize.DefaultOptions()
	r, err := normalize.Apply(snap, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Snapshot.Data["key"]; got != "value" {
		t.Errorf("expected trimmed value, got %q", got)
	}
	if r.Normalized != 1 {
		t.Errorf("expected 1 normalization, got %d", r.Normalized)
	}
}

func TestApply_LowercaseKeys(t *testing.T) {
	snap := makeSnap(map[string]string{
		"HOST": "localhost",
		"PORT": "8080",
	})
	opts := normalize.Options{TrimSpace: false, LowercaseKeys: true, RemoveEmpty: false}
	r, err := normalize.Apply(snap, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Snapshot.Data["host"]; !ok {
		t.Error("expected lowercase key 'host'")
	}
	if _, ok := r.Snapshot.Data["port"]; !ok {
		t.Error("expected lowercase key 'port'")
	}
	if r.Normalized != 2 {
		t.Errorf("expected 2 normalizations, got %d", r.Normalized)
	}
}

func TestApply_RemoveEmptyValues(t *testing.T) {
	snap := makeSnap(map[string]string{
		"present": "yes",
		"empty":   "",
	})
	opts := normalize.Options{TrimSpace: false, LowercaseKeys: false, RemoveEmpty: true}
	r, err := normalize.Apply(snap, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Snapshot.Data["empty"]; ok {
		t.Error("expected empty key to be removed")
	}
	if _, ok := r.Snapshot.Data["present"]; !ok {
		t.Error("expected non-empty key to be preserved")
	}
}

func TestApply_NoChanges(t *testing.T) {
	snap := makeSnap(map[string]string{"key": "value"})
	opts := normalize.Options{}
	r, err := normalize.Apply(snap, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Normalized != 0 {
		t.Errorf("expected 0 normalizations, got %d", r.Normalized)
	}
}

func TestApply_NilDataReturnsError(t *testing.T) {
	snap := snapshot.Snapshot{Environment: "test"}
	_, err := normalize.Apply(snap, normalize.DefaultOptions())
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestRender_NoChanges(t *testing.T) {
	r := normalize.Result{Normalized: 0}
	out := normalize.Render(r)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}

func TestRender_WithChanges(t *testing.T) {
	r := normalize.Result{
		Changes:    []string{`normalized "HOST"`, `normalized "PORT"`},
		Normalized: 2,
	}
	out := normalize.Render(r)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}
