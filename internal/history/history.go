package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a single recorded snapshot event in the history log.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Label     string    `json:"label"`
	FilePath  string    `json:"file_path"`
	SnapshotPath string `json:"snapshot_path"`
}

// Log holds an ordered list of history entries.
type Log struct {
	Entries []Entry `json:"entries"`
}

// Add appends a new entry to the log.
func (l *Log) Add(label, filePath, snapshotPath string) {
	l.Entries = append(l.Entries, Entry{
		Timestamp:    time.Now().UTC(),
		Label:        label,
		FilePath:     filePath,
		SnapshotPath: snapshotPath,
	})
}

// Latest returns the most recent entry for a given file path, or nil if none.
func (l *Log) Latest(filePath string) *Entry {
	for i := len(l.Entries) - 1; i >= 0; i-- {
		if l.Entries[i].FilePath == filePath {
			return &l.Entries[i]
		}
	}
	return nil
}

// Sorted returns entries ordered by timestamp ascending.
func (l *Log) Sorted() []Entry {
	copy := append([]Entry(nil), l.Entries...)
	sort.Slice(copy, func(i, j int) bool {
		return copy[i].Timestamp.Before(copy[j].Timestamp)
	})
	return copy
}

// SaveLog writes the log to disk as JSON.
func SaveLog(path string, log *Log) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("history: mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("history: create: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// LoadLog reads a log from disk. Returns an empty log if the file does not exist.
func LoadLog(path string) (*Log, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return &Log{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("history: open: %w", err)
	}
	defer f.Close()
	var log Log
	if err := json.NewDecoder(f).Decode(&log); err != nil {
		return nil, fmt.Errorf("history: decode: %w", err)
	}
	return &log, nil
}
