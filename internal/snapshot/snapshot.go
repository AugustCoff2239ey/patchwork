package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot holds a point-in-time capture of a configuration file.
type Snapshot struct {
	Environment string            `json:"environment"`
	Timestamp   time.Time         `json:"timestamp"`
	Data        map[string]string `json:"data"`
}

// Capture reads a JSON config file and returns a Snapshot.
func Capture(path, environment string) (Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read %s: %w", path, err)
	}
	var data map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: parse %s: %w", path, err)
	}
	return Snapshot{
		Environment: environment,
		Timestamp:   time.Now().UTC(),
		Data:        data,
	}, nil
}

// Save writes a Snapshot to disk as JSON.
func Save(path string, s Snapshot) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %s: %w", path, err)
	}
	return nil
}

// Load reads a Snapshot from a JSON file on disk.
func Load(path string) (Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read %s: %w", path, err)
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: unmarshal: %w", err)
	}
	return s, nil
}
