package snapshot

import (
	"fmt"
	"sort"
	"time"
)

// ChainEntry represents a single node in a snapshot chain for an environment.
type ChainEntry struct {
	Timestamp time.Time
	Path      string
	Checksum  string
}

// Chain holds an ordered sequence of snapshot entries for a given environment.
type Chain struct {
	Environment string
	Entries     []ChainEntry
}

// BuildChain constructs a Chain from a slice of (path, snapshot) pairs,
// ordered chronologically. Each snapshot must belong to the same environment.
func BuildChain(env string, snaps []Snapshot) (Chain, error) {
	if env == "" {
		return Chain{}, fmt.Errorf("environment must not be empty")
	}
	entries := make([]ChainEntry, 0, len(snaps))
	for _, s := range snaps {
		if s.Environment != env {
			return Chain{}, fmt.Errorf("snapshot environment %q does not match chain environment %q", s.Environment, env)
		}
		entries = append(entries, ChainEntry{
			Timestamp: s.CapturedAt,
			Path:      s.FilePath,
			Checksum:  computeChecksum(s.Data),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})
	return Chain{Environment: env, Entries: entries}, nil
}

// Latest returns the most recent ChainEntry, or an error if the chain is empty.
func (c Chain) Latest() (ChainEntry, error) {
	if len(c.Entries) == 0 {
		return ChainEntry{}, fmt.Errorf("chain for environment %q is empty", c.Environment)
	}
	return c.Entries[len(c.Entries)-1], nil
}

// Len returns the number of entries in the chain.
func (c Chain) Len() int { return len(c.Entries) }

// RenderChain returns a human-readable summary of the chain.
func RenderChain(c Chain) string {
	if len(c.Entries) == 0 {
		return fmt.Sprintf("Chain [%s]: no entries\n", c.Environment)
	}
	out := fmt.Sprintf("Chain [%s] (%d entries):\n", c.Environment, len(c.Entries))
	for i, e := range c.Entries {
		out += fmt.Sprintf("  %d) %s  checksum=%s  path=%s\n",
			i+1, e.Timestamp.Format(time.RFC3339), e.Checksum[:8], e.Path)
	}
	return out
}

// computeChecksum produces a short deterministic checksum string from snapshot data.
func computeChecksum(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := uint64(14695981039346656037)
	for _, k := range keys {
		for _, c := range []byte(k + "=" + data[k] + ";") {
			h ^= uint64(c)
			h *= 1099511628211
		}
	}
	return fmt.Sprintf("%016x", h)
}
