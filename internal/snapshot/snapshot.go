package snapshot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time capture of a configuration file.
type Snapshot struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	Checksum  string    `json:"checksum"`
	Content   string    `json:"content"`
	CapturedAt time.Time `json:"captured_at"`
}

// Capture reads a file and returns a Snapshot of its current state.
func Capture(filePath string) (*Snapshot, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	now := time.Now().UTC()

	return &Snapshot{
		ID:         fmt.Sprintf("%s-%d", checksum[:8], now.UnixNano()),
		FilePath:   filePath,
		Checksum:   checksum,
		Content:    string(data),
		CapturedAt: now,
	}, nil
}

// Save writes the snapshot as JSON to the given destination path.
func (s *Snapshot) Save(destPath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling snapshot: %w", err)
	}
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("writing snapshot to %q: %w", destPath, err)
	}
	return nil
}

// Load reads a snapshot from a JSON file at the given path.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading snapshot file %q: %w", path, err)
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshalling snapshot: %w", err)
	}
	return &s, nil
}
