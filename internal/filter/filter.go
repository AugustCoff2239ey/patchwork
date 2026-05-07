package filter

import (
	"strings"

	"github.com/patchwork/internal/history"
)

// Options holds criteria for filtering history log entries.
type Options struct {
	Environment string
	Since       string // ISO date prefix, e.g. "2024-01"
	HasChanges  bool   // if true, only entries with at least one change
}

// Apply returns a subset of entries matching all non-zero criteria in opts.
func Apply(entries []history.Entry, opts Options) []history.Entry {
	var result []history.Entry
	for _, e := range entries {
		if opts.Environment != "" && !strings.EqualFold(e.Environment, opts.Environment) {
			continue
		}
		if opts.Since != "" && !strings.HasPrefix(e.Timestamp, opts.Since) {
			continue
		}
		if opts.HasChanges && e.ChangeCount == 0 {
			continue
		}
		result = append(result, e)
	}
	return result
}

// Environments returns a deduplicated, sorted list of environment names
// present in the provided entries.
func Environments(entries []history.Entry) []string {
	seen := make(map[string]struct{})
	for _, e := range entries {
		if e.Environment != "" {
			seen[e.Environment] = struct{}{}
		}
	}
	envs := make([]string, 0, len(seen))
	for env := range seen {
		envs = append(envs, env)
	}
	// simple sort
	for i := 0; i < len(envs); i++ {
		for j := i + 1; j < len(envs); j++ {
			if envs[i] > envs[j] {
				envs[i], envs[j] = envs[j], envs[i]
			}
		}
	}
	return envs
}
