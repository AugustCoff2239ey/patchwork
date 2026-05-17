// Package transform applies a sequence of key-value mutations to a Snapshot
// and produces a transformed copy along with a human-readable summary.
package transform

import (
	"fmt"
	"strings"

	"github.com/yourorg/patchwork/internal/snapshot"
)

// OpKind describes the type of transformation.
type OpKind string

const (
	OpSet    OpKind = "set"
	OpDelete OpKind = "delete"
	OpPrefix OpKind = "prefix"
	OpSuffix OpKind = "suffix"
)

// Op is a single transformation instruction.
type Op struct {
	Kind  OpKind `json:"kind"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	Text  string `json:"text,omitempty"`
}

// Result holds the transformed snapshot and a log of applied changes.
type Result struct {
	Snapshot snapshot.Snapshot
	Applied  []string
	Skipped  []string
}

// Apply executes ops against src and returns a Result.
func Apply(src snapshot.Snapshot, ops []Op) (Result, error) {
	if src.Environment == "" {
		return Result{}, fmt.Errorf("transform: source environment must not be empty")
	}
	data := make(map[string]string, len(src.Data))
	for k, v := range src.Data {
		data[k] = v
	}
	var applied, skipped []string
	for _, op := range ops {
		switch op.Kind {
		case OpSet:
			data[op.Key] = op.Value
			applied = append(applied, fmt.Sprintf("set %s=%s", op.Key, op.Value))
		case OpDelete:
			if _, ok := data[op.Key]; ok {
				delete(data, op.Key)
				applied = append(applied, fmt.Sprintf("delete %s", op.Key))
			} else {
				skipped = append(skipped, fmt.Sprintf("delete %s (not found)", op.Key))
			}
		case OpPrefix:
			if v, ok := data[op.Key]; ok {
				data[op.Key] = op.Text + v
				applied = append(applied, fmt.Sprintf("prefix %s with %q", op.Key, op.Text))
			} else {
				skipped = append(skipped, fmt.Sprintf("prefix %s (not found)", op.Key))
			}
		case OpSuffix:
			if v, ok := data[op.Key]; ok {
				data[op.Key] = v + op.Text
				applied = append(applied, fmt.Sprintf("suffix %s with %q", op.Key, op.Text))
			} else {
				skipped = append(skipped, fmt.Sprintf("suffix %s (not found)", op.Key))
			}
		default:
			return Result{}, fmt.Errorf("transform: unknown op kind %q", op.Kind)
		}
	}
	out := snapshot.Snapshot{
		Environment: src.Environment,
		Timestamp:   src.Timestamp,
		Data:        data,
	}
	return Result{Snapshot: out, Applied: applied, Skipped: skipped}, nil
}

// Render returns a human-readable summary of a Result.
func Render(r Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Transform result for environment: %s\n", r.Snapshot.Environment))
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
