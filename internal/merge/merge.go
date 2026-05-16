package merge

import (
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Strategy defines how conflicting keys are resolved during a merge.
type Strategy string

const (
	StrategyOurs   Strategy = "ours"   // keep value from base snapshot
	StrategyTheirs Strategy = "theirs" // keep value from incoming snapshot
	StrategyError  Strategy = "error"  // return error on conflict
)

// Result holds the merged snapshot and metadata about the operation.
type Result struct {
	Snapshot   snapshot.Snapshot
	Conflicts  []Conflict
	MergedKeys int
}

// Conflict records a key that had differing values between base and incoming.
type Conflict struct {
	Key      string
	Base     string
	Incoming string
	Resolved string
}

// Apply merges two snapshots using the given strategy.
// The resulting snapshot inherits the environment and timestamp from base.
func Apply(base, incoming snapshot.Snapshot, strategy Strategy) (Result, error) {
	merged := make(map[string]string)
	var conflicts []Conflict

	for k, v := range base.Data {
		merged[k] = v
	}

	for k, inVal := range incoming.Data {
		baseVal, exists := merged[k]
		if !exists {
			merged[k] = inVal
			continue
		}
		if baseVal == inVal {
			continue
		}
		switch strategy {
		case StrategyOurs:
			conflicts = append(conflicts, Conflict{Key: k, Base: baseVal, Incoming: inVal, Resolved: baseVal})
		case StrategyTheirs:
			merged[k] = inVal
			conflicts = append(conflicts, Conflict{Key: k, Base: baseVal, Incoming: inVal, Resolved: inVal})
		case StrategyError:
			return Result{}, fmt.Errorf("merge conflict on key %q: base=%q incoming=%q", k, baseVal, inVal)
		default:
			return Result{}, fmt.Errorf("unknown merge strategy: %q", strategy)
		}
	}

	result := Result{
		Snapshot:   snapshot.Snapshot{Environment: base.Environment, Timestamp: base.Timestamp, Data: merged},
		Conflicts:  conflicts,
		MergedKeys: len(merged),
	}
	return result, nil
}

// Render returns a human-readable summary of a merge result.
func Render(r Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Merged snapshot: %d keys\n", r.MergedKeys))
	if len(r.Conflicts) == 0 {
		sb.WriteString("No conflicts.\n")
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("Conflicts resolved: %d\n", len(r.Conflicts)))
	sort.Slice(r.Conflicts, func(i, j int) bool { return r.Conflicts[i].Key < r.Conflicts[j].Key })
	for _, c := range r.Conflicts {
		sb.WriteString(fmt.Sprintf("  [conflict] %s: base=%q incoming=%q -> resolved=%q\n", c.Key, c.Base, c.Incoming, c.Resolved))
	}
	return sb.String()
}
