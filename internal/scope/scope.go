package scope

import (
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Scope defines a named subset of configuration keys.
type Scope struct {
	Name        string   `json:"name"`
	Environment string   `json:"environment"`
	Keys        []string `json:"keys"`
}

// Result holds the filtered snapshot data for a scope.
type Result struct {
	Scope   Scope
	Matched map[string]string
	Missing []string
}

// Apply filters a snapshot to only the keys defined in the scope.
func Apply(snap snapshot.Snapshot, sc Scope) (Result, error) {
	if sc.Name == "" {
		return Result{}, fmt.Errorf("scope name must not be empty")
	}
	if sc.Environment != snap.Environment {
		return Result{}, fmt.Errorf("scope environment %q does not match snapshot environment %q", sc.Environment, snap.Environment)
	}

	matched := make(map[string]string)
	var missing []string

	for _, k := range sc.Keys {
		if v, ok := snap.Data[k]; ok {
			matched[k] = v
		} else {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)

	return Result{Scope: sc, Matched: matched, Missing: missing}, nil
}

// Render returns a human-readable representation of a scope result.
func Render(r Result) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Scope: %s (%s)\n", r.Scope.Name, r.Scope.Environment)
	fmt.Fprintf(&sb, "Matched: %d / %d keys\n", len(r.Matched), len(r.Scope.Keys))

	keys := make([]string, 0, len(r.Matched))
	for k := range r.Matched {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&sb, "  %s = %s\n", k, r.Matched[k])
	}

	if len(r.Missing) > 0 {
		fmt.Fprintf(&sb, "Missing keys: %s\n", strings.Join(r.Missing, ", "))
	}
	return sb.String()
}
