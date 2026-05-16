package template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Template holds a named set of default key-value pairs for an environment.
type Template struct {
	Name        string            `json:"name"`
	Environment string            `json:"environment"`
	Defaults    map[string]string `json:"defaults"`
}

// ApplyResult describes the outcome of applying a template to a snapshot.
type ApplyResult struct {
	Template string
	Added    []string
	Skipped  []string
}

// Apply merges template defaults into the snapshot, skipping keys that already
// exist. Returns an ApplyResult summarising what changed.
func Apply(t Template, snap snapshot.Snapshot, overwrite bool) (snapshot.Snapshot, ApplyResult, error) {
	if t.Name == "" {
		return snap, ApplyResult{}, fmt.Errorf("template name must not be empty")
	}
	if snap.Data == nil {
		snap.Data = make(map[string]string)
	}

	result := ApplyResult{Template: t.Name}

	keys := make([]string, 0, len(t.Defaults))
	for k := range t.Defaults {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, exists := snap.Data[k]
		if exists && !overwrite {
			result.Skipped = append(result.Skipped, k)
			continue
		}
		snap.Data[k] = t.Defaults[k]
		result.Added = append(result.Added, k)
	}

	return snap, result, nil
}

// Render returns a human-readable summary of an ApplyResult.
func Render(r ApplyResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Template: %s\n", r.Template)
	fmt.Fprintf(&sb, "  Added:   %d key(s)\n", len(r.Added))
	for _, k := range r.Added {
		fmt.Fprintf(&sb, "    + %s\n", k)
	}
	fmt.Fprintf(&sb, "  Skipped: %d key(s)\n", len(r.Skipped))
	for _, k := range r.Skipped {
		fmt.Fprintf(&sb, "    ~ %s\n", k)
	}
	return sb.String()
}
