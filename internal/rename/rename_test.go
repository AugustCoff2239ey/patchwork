package rename

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		CapturedAt:  time.Now(),
		Data:        data,
	}
}

func TestApply_RenamesExistingKey(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"db_host": "localhost", "port": "5432"})
	updated, result := Apply(snap, "db_host", "database_host")

	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.OldKey != "db_host" || result.NewKey != "database_host" {
		t.Errorf("unexpected keys: %s -> %s", result.OldKey, result.NewKey)
	}
	if result.Value != "localhost" {
		t.Errorf("expected value 'localhost', got %s", result.Value)
	}
	if _, ok := updated.Data["db_host"]; ok {
		t.Error("old key should not exist in updated snapshot")
	}
	if v, ok := updated.Data["database_host"]; !ok || v != "localhost" {
		t.Errorf("expected new key 'database_host'='localhost', got %s", v)
	}
	if updated.Data["port"] != "5432" {
		t.Error("unrelated keys should be preserved")
	}
}

func TestApply_MissingKey_ReturnsFalse(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"host": "example.com"})
	updated, result := Apply(snap, "missing_key", "new_key")

	if result.Found {
		t.Fatal("expected Found=false for missing key")
	}
	if len(updated.Data) != len(snap.Data) {
		t.Error("snapshot should be unchanged when key is missing")
	}
}

func TestApply_PreservesEnvironmentAndTimestamp(t *testing.T) {
	snap := makeSnap("dev", map[string]string{"key": "value"})
	updated, _ := Apply(snap, "key", "renamed_key")

	if updated.Environment != snap.Environment {
		t.Errorf("environment mismatch: %s != %s", updated.Environment, snap.Environment)
	}
	if !updated.CapturedAt.Equal(snap.CapturedAt) {
		t.Error("captured_at should be preserved")
	}
}

func TestRender_FoundKey(t *testing.T) {
	r := Result{Environment: "prod", OldKey: "db_host", NewKey: "database_host", Value: "localhost", Found: true}
	out := Render(r)

	if !strings.Contains(out, "prod") {
		t.Error("expected environment in output")
	}
	if !strings.Contains(out, "db_host") || !strings.Contains(out, "database_host") {
		t.Error("expected old and new key names in output")
	}
	if !strings.Contains(out, "localhost") {
		t.Error("expected value in output")
	}
}

func TestRender_NotFound(t *testing.T) {
	r := Result{Environment: "dev", OldKey: "ghost", NewKey: "spirit", Found: false}
	out := Render(r)

	if !strings.Contains(out, "not found") {
		t.Error("expected 'not found' message when key is missing")
	}
}
