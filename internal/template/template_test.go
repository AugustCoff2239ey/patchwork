package template

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: "staging",
		CapturedAt:  time.Now(),
		Data:        data,
	}
}

func TestApply_AddsNewKeys(t *testing.T) {
	tmpl := Template{
		Name:        "base",
		Environment: "staging",
		Defaults:    map[string]string{"LOG_LEVEL": "info", "TIMEOUT": "30"},
	}
	snap := makeSnap(map[string]string{})
	out, result, err := Apply(tmpl, snap, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Added) != 2 {
		t.Errorf("expected 2 added, got %d", len(result.Added))
	}
	if out.Data["LOG_LEVEL"] != "info" {
		t.Errorf("expected LOG_LEVEL=info, got %s", out.Data["LOG_LEVEL"])
	}
}

func TestApply_SkipsExistingKeysWithoutOverwrite(t *testing.T) {
	tmpl := Template{
		Name:     "base",
		Defaults: map[string]string{"LOG_LEVEL": "info"},
	}
	snap := makeSnap(map[string]string{"LOG_LEVEL": "debug"})
	out, result, err := Apply(tmpl, snap, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
	if out.Data["LOG_LEVEL"] != "debug" {
		t.Errorf("expected LOG_LEVEL=debug, got %s", out.Data["LOG_LEVEL"])
	}
}

func TestApply_OverwriteReplacesExisting(t *testing.T) {
	tmpl := Template{
		Name:     "base",
		Defaults: map[string]string{"LOG_LEVEL": "warn"},
	}
	snap := makeSnap(map[string]string{"LOG_LEVEL": "debug"})
	out, result, err := Apply(tmpl, snap, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(result.Added))
	}
	if out.Data["LOG_LEVEL"] != "warn" {
		t.Errorf("expected LOG_LEVEL=warn, got %s", out.Data["LOG_LEVEL"])
	}
}

func TestApply_EmptyTemplateName_ReturnsError(t *testing.T) {
	tmpl := Template{Defaults: map[string]string{"K": "v"}}
	_, _, err := Apply(tmpl, makeSnap(nil), false)
	if err == nil {
		t.Error("expected error for empty template name")
	}
}

func TestRender_ContainsTemplateName(t *testing.T) {
	r := ApplyResult{Template: "mytemplate", Added: []string{"A"}, Skipped: []string{"B"}}
	out := Render(r)
	if !strings.Contains(out, "mytemplate") {
		t.Error("expected template name in render output")
	}
	if !strings.Contains(out, "+ A") {
		t.Error("expected added key in render output")
	}
	if !strings.Contains(out, "~ B") {
		t.Error("expected skipped key in render output")
	}
}
