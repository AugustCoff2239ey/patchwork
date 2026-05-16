package index

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/patchwork/internal/history"
)

// Entry represents a searchable index record for a snapshot key.
type Entry struct {
	Environment string `json:"environment"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	SnapshotID  string `json:"snapshot_id"`
	Timestamp   string `json:"timestamp"`
}

// Index holds all indexed entries.
type Index struct {
	Entries []Entry `json:"entries"`
}

// Build constructs an Index from a history log by scanning all snapshot data.
func Build(log []history.Entry) Index {
	var entries []Entry
	for _, h := range log {
		for k, v := range h.Snapshot.Data {
			entries = append(entries, Entry{
				Environment: h.Snapshot.Environment,
				Key:         k,
				Value:       v,
				SnapshotID:  h.ID,
				Timestamp:   h.Snapshot.Timestamp,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Environment != entries[j].Environment {
			return entries[i].Environment < entries[j].Environment
		}
		return entries[i].Key < entries[j].Key
	})
	return Index{Entries: entries}
}

// Search returns entries matching the given key substring and optional environment.
func Search(idx Index, query, env string) []Entry {
	var results []Entry
	for _, e := range idx.Entries {
		if env != "" && e.Environment != env {
			continue
		}
		if strings.Contains(e.Key, query) || strings.Contains(e.Value, query) {
			results = append(results, e)
		}
	}
	return results
}

// Save writes the index to disk as JSON.
func Save(idx Index, path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("index: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads an index from disk.
func Load(path string) (Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Index{}, nil
		}
		return Index{}, fmt.Errorf("index: read: %w", err)
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return Index{}, fmt.Errorf("index: unmarshal: %w", err)
	}
	return idx, nil
}

// Render returns a human-readable summary of search results.
func Render(results []Entry) string {
	if len(results) == 0 {
		return "No matching entries found.\n"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d result(s):\n\n", len(results)))
	for _, e := range results {
		sb.WriteString(fmt.Sprintf("  [%s] %s = %s\n    snapshot: %s @ %s\n",
			e.Environment, e.Key, e.Value, e.SnapshotID, e.Timestamp))
	}
	return sb.String()
}
