package snapshot

import (
	"fmt"
	"sort"
	"time"
)

// DiffSnapshot represents the delta between two snapshots at a point in time.
type DiffSnapshot struct {
	Environment string            `json:"environment"`
	From        time.Time         `json:"from"`
	To          time.Time         `json:"to"`
	Added       map[string]string `json:"added"`
	Removed     map[string]string `json:"removed"`
	Modified    map[string]string `json:"modified"` // key -> new value
}

// Delta computes a DiffSnapshot between two Snapshot values.
func Delta(before, after Snapshot) (DiffSnapshot, error) {
	if before.Environment != after.Environment {
		return DiffSnapshot{}, fmt.Errorf("environment mismatch: %q vs %q", before.Environment, after.Environment)
	}

	d := DiffSnapshot{
		Environment: after.Environment,
		From:        before.Timestamp,
		To:          after.Timestamp,
		Added:       make(map[string]string),
		Removed:     make(map[string]string),
		Modified:    make(map[string]string),
	}

	for k, v := range after.Data {
		if old, ok := before.Data[k]; !ok {
			d.Added[k] = v
		} else if old != v {
			d.Modified[k] = v
		}
	}

	for k, v := range before.Data {
		if _, ok := after.Data[k]; !ok {
			d.Removed[k] = v
		}
	}

	return d, nil
}

// Keys returns a sorted list of all changed keys across added, removed, and modified.
func (d DiffSnapshot) Keys() []string {
	seen := make(map[string]struct{})
	for k := range d.Added {
		seen[k] = struct{}{}
	}
	for k := range d.Removed {
		seen[k] = struct{}{}
	}
	for k := range d.Modified {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// IsEmpty reports whether the DiffSnapshot contains no changes.
func (d DiffSnapshot) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Modified) == 0
}
