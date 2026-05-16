package archive

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/patchwork/internal/history"
)

// Entry represents a single archived snapshot bundle.
type Entry struct {
	ID          string    `json:"id"`
	Environment string    `json:"environment"`
	ArchivedAt  time.Time `json:"archived_at"`
	SnapshotRef string    `json:"snapshot_ref"`
	Note        string    `json:"note,omitempty"`
}

// Archive holds a collection of archived entries.
type Archive struct {
	Entries []Entry `json:"entries"`
}

// Build creates an Archive from history log entries, optionally filtered by environment.
func Build(log history.Log, env string) Archive {
	a := Archive{}
	for _, e := range log.Entries {
		if env != "" && e.Environment != env {
			continue
		}
		a.Entries = append(a.Entries, Entry{
			ID:          e.ID,
			Environment: e.Environment,
			ArchivedAt:  time.Now().UTC(),
			SnapshotRef: e.SnapshotPath,
			Note:        fmt.Sprintf("%d change(s)", len(e.Changes)),
		})
	}
	sort.Slice(a.Entries, func(i, j int) bool {
		return a.Entries[i].ArchivedAt.Before(a.Entries[j].ArchivedAt)
	})
	return a
}

// Save writes an Archive to disk as JSON.
func Save(path string, a Archive) error {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("archive: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads an Archive from a JSON file.
func Load(path string) (Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Archive{}, nil
		}
		return Archive{}, fmt.Errorf("archive: read: %w", err)
	}
	var a Archive
	if err := json.Unmarshal(data, &a); err != nil {
		return Archive{}, fmt.Errorf("archive: unmarshal: %w", err)
	}
	return a, nil
}

// Render returns a human-readable summary of the archive.
func Render(a Archive) string {
	if len(a.Entries) == 0 {
		return "archive: no entries\n"
	}
	out := fmt.Sprintf("Archive (%d entries)\n", len(a.Entries))
	for _, e := range a.Entries {
		out += fmt.Sprintf("  [%s] env=%s ref=%s note=%q\n",
			e.ID, e.Environment, e.SnapshotRef, e.Note)
	}
	return out
}
