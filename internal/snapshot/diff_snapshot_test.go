package snapshot

import (
	"testing"
	"time"
)

func baseSnap(env string, data map[string]string) Snapshot {
	return Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
}

func TestDelta_DetectsAddedKeys(t *testing.T) {
	before := baseSnap("prod", map[string]string{"a": "1"})
	after := baseSnap("prod", map[string]string{"a": "1", "b": "2"})

	d, err := Delta(before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := d.Added["b"]; !ok {
		t.Error("expected key 'b' in Added")
	}
	if len(d.Removed) != 0 || len(d.Modified) != 0 {
		t.Error("expected no removed or modified keys")
	}
}

func TestDelta_DetectsRemovedKeys(t *testing.T) {
	before := baseSnap("prod", map[string]string{"a": "1", "b": "2"})
	after := baseSnap("prod", map[string]string{"a": "1"})

	d, err := Delta(before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := d.Removed["b"]; !ok {
		t.Error("expected key 'b' in Removed")
	}
}

func TestDelta_DetectsModifiedKeys(t *testing.T) {
	before := baseSnap("prod", map[string]string{"a": "old"})
	after := baseSnap("prod", map[string]string{"a": "new"})

	d, err := Delta(before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := d.Modified["a"]; !ok || v != "new" {
		t.Errorf("expected Modified[a]=new, got %q", v)
	}
}

func TestDelta_EnvironmentMismatch(t *testing.T) {
	before := baseSnap("prod", map[string]string{})
	after := baseSnap("staging", map[string]string{})

	_, err := Delta(before, after)
	if err == nil {
		t.Error("expected error for environment mismatch")
	}
}

func TestDelta_IsEmpty_WhenNoChanges(t *testing.T) {
	snap := baseSnap("prod", map[string]string{"a": "1"})
	d, err := Delta(snap, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.IsEmpty() {
		t.Error("expected empty diff")
	}
}

func TestDelta_Keys_ReturnsSorted(t *testing.T) {
	before := baseSnap("prod", map[string]string{"b": "1", "c": "old"})
	after := baseSnap("prod", map[string]string{"a": "new", "c": "new"})

	d, _ := Delta(before, after)
	keys := d.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("unexpected key order: %v", keys)
	}
}
