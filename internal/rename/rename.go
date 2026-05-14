package rename

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Result holds the outcome of a key rename operation.
type Result struct {
	Environment string
	OldKey      string
	NewKey      string
	Value       string
	Found       bool
}

// Apply renames oldKey to newKey in the given snapshot, returning a Result.
// If oldKey does not exist, Found is false and the snapshot is unchanged.
func Apply(snap snapshot.Snapshot, oldKey, newKey string) (snapshot.Snapshot, Result) {
	result := Result{
		Environment: snap.Environment,
		OldKey:      oldKey,
		NewKey:      newKey,
	}

	val, ok := snap.Data[oldKey]
	if !ok {
		return snap, result
	}

	updated := snapshot.Snapshot{
		Environment: snap.Environment,
		CapturedAt:  snap.CapturedAt,
		Data:        make(map[string]string, len(snap.Data)),
	}
	for k, v := range snap.Data {
		if k == oldKey {
			updated.Data[newKey] = v
		} else {
			updated.Data[k] = v
		}
	}

	result.Value = val
	result.Found = true
	return updated, result
}

// Render formats a Result as a human-readable string.
func Render(r Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Environment : %s\n", r.Environment))
	if !r.Found {
		sb.WriteString(fmt.Sprintf("Key '%s' not found — no changes made.\n", r.OldKey))
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("Renamed     : %s  →  %s\n", r.OldKey, r.NewKey))
	sb.WriteString(fmt.Sprintf("Value       : %s\n", r.Value))
	return sb.String()
}
