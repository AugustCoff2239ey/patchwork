package patch

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Op represents a single patch operation.
type Op struct {
	Action string // "set", "delete", "rename"
	Key    string
	Value  string
	NewKey string // used for rename
}

// Plan holds a list of patch operations to apply.
type Plan struct {
	Environment string
	Ops         []Op
}

// Result describes the outcome of applying a plan.
type Result struct {
	Applied []string
	Skipped []string
	Snap    snapshot.Snapshot
}

// Apply executes the patch plan against the given snapshot.
func Apply(plan Plan, snap snapshot.Snapshot) (Result, error) {
	if plan.Environment == "" {
		return Result{}, fmt.Errorf("patch: environment must not be empty")
	}

	data := make(map[string]string, len(snap.Data))
	for k, v := range snap.Data {
		data[k] = v
	}

	var applied, skipped []string

	for _, op := range plan.Ops {
		switch strings.ToLower(op.Action) {
		case "set":
			data[op.Key] = op.Value
			applied = append(applied, fmt.Sprintf("set %s", op.Key))
		case "delete":
			if _, ok := data[op.Key]; ok {
				delete(data, op.Key)
				applied = append(applied, fmt.Sprintf("delete %s", op.Key))
			} else {
				skipped = append(skipped, fmt.Sprintf("delete %s (not found)", op.Key))
			}
		case "rename":
			if val, ok := data[op.Key]; ok {
				data[op.NewKey] = val
				delete(data, op.Key)
				applied = append(applied, fmt.Sprintf("rename %s -> %s", op.Key, op.NewKey))
			} else {
				skipped = append(skipped, fmt.Sprintf("rename %s (not found)", op.Key))
			}
		default:
			skipped = append(skipped, fmt.Sprintf("unknown action %q on key %s", op.Action, op.Key))
		}
	}

	result := Result{
		Applied: applied,
		Skipped: skipped,
		Snap: snapshot.Snapshot{
			Environment: plan.Environment,
			Timestamp:   snap.Timestamp,
			Data:        data,
		},
	}
	return result, nil
}

// Render returns a human-readable summary of the patch result.
func Render(r Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Patch result for environment: %s\n", r.Snap.Environment))
	sb.WriteString(fmt.Sprintf("  Applied : %d\n", len(r.Applied)))
	for _, a := range r.Applied {
		sb.WriteString(fmt.Sprintf("    + %s\n", a))
	}
	sb.WriteString(fmt.Sprintf("  Skipped : %d\n", len(r.Skipped)))
	for _, s := range r.Skipped {
		sb.WriteString(fmt.Sprintf("    - %s\n", s))
	}
	return sb.String()
}
