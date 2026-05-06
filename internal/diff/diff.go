package diff

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Change represents a single configuration change between two snapshots.
type Change struct {
	Key      string
	OldValue string
	NewValue string
	Type     ChangeType
}

// ChangeType describes the nature of a configuration change.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
)

// Result holds the full diff between two snapshots.
type Result struct {
	From    string
	To      string
	Changes []Change
}

// Compare computes the diff between two snapshots and returns a Result.
func Compare(from, to *snapshot.Snapshot) *Result {
	result := &Result{
		From: from.Timestamp.String(),
		To:   to.Timestamp.String(),
	}

	for key, newVal := range to.Entries {
		oldVal, exists := from.Entries[key]
		if !exists {
			result.Changes = append(result.Changes, Change{Key: key, NewValue: newVal, Type: Added})
		} else if oldVal != newVal {
			result.Changes = append(result.Changes, Change{Key: key, OldValue: oldVal, NewValue: newVal, Type: Modified})
		}
	}

	for key, oldVal := range from.Entries {
		if _, exists := to.Entries[key]; !exists {
			result.Changes = append(result.Changes, Change{Key: key, OldValue: oldVal, Type: Removed})
		}
	}

	return result
}

// Format returns a human-readable string representation of the diff result.
func Format(r *Result) string {
	if len(r.Changes) == 0 {
		return "No changes detected."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Diff from %s to %s:\n", r.From, r.To)
	for _, c := range r.Changes {
		switch c.Type {
		case Added:
			fmt.Fprintf(&sb, "  + %s = %q\n", c.Key, c.NewValue)
		case Removed:
			fmt.Fprintf(&sb, "  - %s = %q\n", c.Key, c.OldValue)
		case Modified:
			fmt.Fprintf(&sb, "  ~ %s: %q -> %q\n", c.Key, c.OldValue, c.NewValue)
		}
	}
	return sb.String()
}
