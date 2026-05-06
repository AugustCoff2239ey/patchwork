package diff_test

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/snapshot"
)

func makeSnapshot(entries map[string]string) *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Timestamp: time.Now(),
		Entries:   entries,
	}
}

func TestCompare_DetectsAddedKeys(t *testing.T) {
	from := makeSnapshot(map[string]string{"a": "1"})
	to := makeSnapshot(map[string]string{"a": "1", "b": "2"})

	result := diff.Compare(from, to)

	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0].Type != diff.Added || result.Changes[0].Key != "b" {
		t.Errorf("expected added key 'b', got %+v", result.Changes[0])
	}
}

func TestCompare_DetectsRemovedKeys(t *testing.T) {
	from := makeSnapshot(map[string]string{"a": "1", "b": "2"})
	to := makeSnapshot(map[string]string{"a": "1"})

	result := diff.Compare(from, to)

	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0].Type != diff.Removed || result.Changes[0].Key != "b" {
		t.Errorf("expected removed key 'b', got %+v", result.Changes[0])
	}
}

func TestCompare_DetectsModifiedKeys(t *testing.T) {
	from := makeSnapshot(map[string]string{"a": "old"})
	to := makeSnapshot(map[string]string{"a": "new"})

	result := diff.Compare(from, to)

	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Type != diff.Modified || c.OldValue != "old" || c.NewValue != "new" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	entries := map[string]string{"x": "1", "y": "2"}
	from := makeSnapshot(entries)
	to := makeSnapshot(entries)

	result := diff.Compare(from, to)

	if len(result.Changes) != 0 {
		t.Errorf("expected no changes, got %d", len(result.Changes))
	}
}

func TestFormat_NoChanges(t *testing.T) {
	r := &diff.Result{}
	out := diff.Format(r)
	if out != "No changes detected." {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormat_WithChanges(t *testing.T) {
	from := makeSnapshot(map[string]string{"a": "1"})
	to := makeSnapshot(map[string]string{"a": "2", "b": "3"})

	r := diff.Compare(from, to)
	out := diff.Format(r)

	if !strings.Contains(out, "+ b") && !strings.Contains(out, "~ a") {
		t.Errorf("formatted output missing expected lines:\n%s", out)
	}
}
