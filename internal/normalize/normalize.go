package normalize

import (
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Options controls normalization behavior.
type Options struct {
	TrimSpace    bool
	LowercaseKeys bool
	RemoveEmpty  bool
}

// DefaultOptions returns sensible normalization defaults.
func DefaultOptions() Options {
	return Options{
		TrimSpace:    true,
		LowercaseKeys: false,
		RemoveEmpty:  false,
	}
}

// Result holds the outcome of a normalization pass.
type Result struct {
	Snapshot  snapshot.Snapshot
	Changes   []string
	Normalized int
}

// Apply normalizes a snapshot according to the given options.
// It returns a new snapshot with the transformations applied and a Result
// describing what changed.
func Apply(snap snapshot.Snapshot, opts Options) (Result, error) {
	if snap.Data == nil {
		return Result{}, fmt.Errorf("normalize: snapshot data is nil")
	}

	newData := make(map[string]string, len(snap.Data))
	var changes []string

	for k, v := range snap.Data {
		newKey := k
		newVal := v

		if opts.TrimSpace {
			newKey = strings.TrimSpace(newKey)
			newVal = strings.TrimSpace(newVal)
		}

		if opts.LowercaseKeys {
			newKey = strings.ToLower(newKey)
		}

		if opts.RemoveEmpty && newVal == "" {
			changes = append(changes, fmt.Sprintf("removed empty key %q", k))
			continue
		}

		if newKey != k || newVal != v {
			changes = append(changes, fmt.Sprintf("normalized %q", k))
		}

		newData[newKey] = newVal
	}

	sort.Strings(changes)

	out := snapshot.Snapshot{
		Environment: snap.Environment,
		Timestamp:   snap.Timestamp,
		Data:        newData,
	}

	return Result{
		Snapshot:   out,
		Changes:    changes,
		Normalized: len(changes),
	}, nil
}

// Render returns a human-readable summary of the normalization result.
func Render(r Result) string {
	if r.Normalized == 0 {
		return "normalize: no changes required\n"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "normalize: %d change(s) applied\n", r.Normalized)
	for _, c := range r.Changes {
		fmt.Fprintf(&sb, "  - %s\n", c)
	}
	return sb.String()
}
