package summary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/history"
)

// Stats holds aggregated change statistics across history entries.
type Stats struct {
	Environment  string
	TotalEntries int
	TotalAdded   int
	TotalRemoved int
	TotalModified int
	MostChanged  string
}

// Build computes a Stats summary for a given environment from log entries.
func Build(entries []history.Entry, env string) Stats {
	s := Stats{Environment: env}

	keyCounts := map[string]int{}

	for _, e := range entries {
		if env != "" && e.Environment != env {
			continue
		}
		s.TotalEntries++
		for _, c := range e.Changes {
			switch c.Type {
			case "added":
				s.TotalAdded++
			case "removed":
				s.TotalRemoved++
			case "modified":
				s.TotalModified++
			}
			keyCounts[c.Key]++
		}
	}

	s.MostChanged = topKey(keyCounts)
	return s
}

// Render formats a Stats value as a human-readable string.
func Render(s Stats) string {
	var b strings.Builder
	label := s.Environment
	if label == "" {
		label = "all environments"
	}
	fmt.Fprintf(&b, "Summary (%s)\n", label)
	fmt.Fprintf(&b, "  Entries   : %d\n", s.TotalEntries)
	fmt.Fprintf(&b, "  Added     : %d\n", s.TotalAdded)
	fmt.Fprintf(&b, "  Removed   : %d\n", s.TotalRemoved)
	fmt.Fprintf(&b, "  Modified  : %d\n", s.TotalModified)
	if s.MostChanged != "" {
		fmt.Fprintf(&b, "  Top Key   : %s\n", s.MostChanged)
	}
	return b.String()
}

func topKey(counts map[string]int) string {
	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})
	if len(pairs) == 0 {
		return ""
	}
	return pairs[0].key
}
