package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time capture of configuration entries.
type Snapshot struct {
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
	Entries   map[string]string `json:"entries"`
}

// Capture reads a config file at the given path and returns a Snapshot.
func Capture(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("capture: reading file %q: %w", path, err)
	}

	entries := make(map[string]string)
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("capture: parsing file %q: %w", path, err)
	}

	return &Snapshot{
		Timestamp: time.Now().UTC(),
		Source:    path,
		Entries:   entries,
	}, nil
}

// Save writes a snapshot to the given path as JSON.
func Save(s *Snapshot, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshalling snapshot: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("save: writing file %q: %w", path, err)
	}
	return nil
}

// Load reads a previously saved snapshot from disk.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load: reading file %q: %w", path, err)
	}

	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("load: parsing snapshot %q: %w", path, err)
	}
	return &s, nil
}
