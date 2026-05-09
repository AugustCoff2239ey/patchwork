// Package compare provides multi-snapshot comparison across environments,
// allowing patchwork to detect configuration divergence between environments
// (e.g. staging vs production) at a given point in time.
package compare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/user/patchwork/internal/snapshot"
)

// Result holds the outcome of comparing two environment snapshots.
type Result struct {
	BaseEnv    string
	TargetEnv  string
	OnlyInBase []string // keys present only in base
	OnlyInTarget []string // keys present only in target
	Diverged   []Divergence // keys present in both but with different values
}

// Divergence describes a single key whose value differs between environments.
type Divergence struct {
	Key         string
	BaseValue   string
	TargetValue string
}

// Environments compares two snapshots from different environments and returns
// a Result describing how they diverge. Keys present in both snapshots are
// compared by value; missing keys are reported separately.
func Environments(base, target snapshot.Snapshot) Result {
	result := Result{
		BaseEnv:   base.Environment,
		TargetEnv: target.Environment,
	}

	baseKeys := keySet(base.Data)
	targetKeys := keySet(target.Data)

	// Keys only in base
	for k := range baseKeys {
		if _, ok := targetKeys[k]; !ok {
			result.OnlyInBase = append(result.OnlyInBase, k)
		}
	}

	// Keys only in target
	for k := range targetKeys {
		if _, ok := baseKeys[k]; !ok {
			result.OnlyInTarget = append(result.OnlyInTarget, k)
		}
	}

	// Keys in both — check for value divergence
	for k := range baseKeys {
		if tv, ok := target.Data[k]; ok {
			if base.Data[k] != tv {
				result.Diverged = append(result.Diverged, Divergence{
					Key:         k,
					BaseValue:   base.Data[k],
					TargetValue: tv,
				})
			}
		}
	}

	sort.Strings(result.OnlyInBase)
	sort.Strings(result.OnlyInTarget)
	sort.Slice(result.Diverged, func(i, j int) bool {
		return result.Diverged[i].Key < result.Diverged[j].Key
	})

	return result
}

// HasDrift returns true if any differences were found between the two snapshots.
func (r Result) HasDrift() bool {
	return len(r.OnlyInBase) > 0 || len(r.OnlyInTarget) > 0 || len(r.Diverged) > 0
}

// Render formats the comparison result as a human-readable string.
func Render(r Result) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Environment comparison: %s → %s\n", r.BaseEnv, r.TargetEnv)

	if !r.HasDrift() {
		sb.WriteString("  No divergence detected.\n")
		return sb.String()
	}

	if len(r.OnlyInBase) > 0 {
		fmt.Fprintf(&sb, "\n  Only in %s:\n", r.BaseEnv)
		for _, k := range r.OnlyInBase {
			fmt.Fprintf(&sb, "    - %s\n", k)
		}
	}

	if len(r.OnlyInTarget) > 0 {
		fmt.Fprintf(&sb, "\n  Only in %s:\n", r.TargetEnv)
		for _, k := range r.OnlyInTarget {
			fmt.Fprintf(&sb, "    + %s\n", k)
		}
	}

	if len(r.Diverged) > 0 {
		sb.WriteString("\n  Diverged values:\n")
		for _, d := range r.Diverged {
			fmt.Fprintf(&sb, "    ~ %s\n      %s: %q\n      %s: %q\n",
				d.Key, r.BaseEnv, d.BaseValue, r.TargetEnv, d.TargetValue)
		}
	}

	return sb.String()
}

// keySet builds a set of keys from a data map for quick lookup.
func keySet(data map[string]string) map[string]struct{} {
	set := make(map[string]struct{}, len(data))
	for k := range data {
		set[k] = struct{}{}
	}
	return set
}
