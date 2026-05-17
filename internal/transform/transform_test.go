package transform_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/patchwork/internal/snapshot"
	"github.com/yourorg/patchwork/internal/transform"
)

func makeSnap(data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: "staging",
		Timestamp:   time.Now().UTC(),
		Data:        data,
	}
}

func TestApply_SetAddsOrUpdatesKey(t *testing.T) {
	src := makeSnap(map[string]string{"host": "localhost"})
	ops := []transform.Op{
		{Kind: transform.OpSet, Key: "port", Value: "8080"},
		{Kind: transform.OpSet, Key: "host", Value: "prod.example.com"},
	}
	r, err := transform.Apply(src, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snapshot.Data["port"] != "8080" {
		t.Errorf("expected port=8080, got %s", r.Snapshot.Data["port"])
	}
	if r.Snapshot.Data["host"] != "prod.example.com" {
		t.Errorf("expected host=prod.example.com, got %s", r.Snapshot.Data["host"])
	}
	if len(r.Applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(r.Applied))
	}
}

func TestApply_DeleteRemovesKey(t *testing.T) {
	src := makeSnap(map[string]string{"a": "1", "b": "2"})
	ops := []transform.Op{{Kind: transform.OpDelete, Key: "a"}}
	r, err := transform.Apply(src, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Snapshot.Data["a"]; ok {
		t.Error("expected key 'a' to be deleted")
	}
	if len(r.Applied) != 1 {
		t.Errorf("expected 1 applied, got %d", len(r.Applied))
	}
}

func TestApply_DeleteMissingKeyIsSkipped(t *testing.T) {
	src := makeSnap(map[string]string{"b": "2"})
	ops := []transform.Op{{Kind: transform.OpDelete, Key: "missing"}}
	r, err := transform.Apply(src, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(r.Skipped))
	}
}

func TestApply_PrefixAndSuffix(t *testing.T) {
	src := makeSnap(map[string]string{"url": "example.com"})
	ops := []transform.Op{
		{Kind: transform.OpPrefix, Key: "url", Text: "https://"},
		{Kind: transform.OpSuffix, Key: "url", Text: "/api"},
	}
	r, err := transform.Apply(src, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Snapshot.Data["url"] != "https://example.com/api" {
		t.Errorf("unexpected url: %s", r.Snapshot.Data["url"])
	}
}

func TestApply_EmptyEnvironmentReturnsError(t *testing.T) {
	src := snapshot.Snapshot{Data: map[string]string{"k": "v"}}
	_, err := transform.Apply(src, nil)
	if err == nil {
		t.Error("expected error for empty environment")
	}
}

func TestApply_UnknownOpReturnsError(t *testing.T) {
	src := makeSnap(map[string]string{})
	ops := []transform.Op{{Kind: "explode", Key: "x"}}
	_, err := transform.Apply(src, ops)
	if err == nil {
		t.Error("expected error for unknown op kind")
	}
}

func TestRender_ContainsEnvironment(t *testing.T) {
	src := makeSnap(map[string]string{"k": "v"})
	r, _ := transform.Apply(src, []transform.Op{{Kind: transform.OpSet, Key: "k", Value: "new"}})
	out := transform.Render(r)
	if !strings.Contains(out, "staging") {
		t.Error("expected render to contain environment name")
	}
	if !strings.Contains(out, "Applied") {
		t.Error("expected render to contain 'Applied'")
	}
}
