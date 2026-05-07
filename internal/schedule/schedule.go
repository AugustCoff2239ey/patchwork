package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry represents a scheduled capture job for a given environment and config file.
type Entry struct {
	Environment string        `json:"environment"`
	ConfigPath  string        `json:"config_path"`
	Interval    time.Duration `json:"interval_ns"`
	LastRun     time.Time     `json:"last_run,omitempty"`
}

// Schedule holds a collection of scheduled entries.
type Schedule struct {
	Entries []Entry `json:"entries"`
}

// IsDue returns true if the entry is due to run based on the current time.
func (e *Entry) IsDue(now time.Time) bool {
	if e.LastRun.IsZero() {
		return true
	}
	return now.Sub(e.LastRun) >= e.Interval
}

// MarkRun updates the LastRun timestamp to now.
func (e *Entry) MarkRun(now time.Time) {
	e.LastRun = now
}

// Save writes the schedule to a JSON file at the given path.
func Save(path string, s *Schedule) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("schedule: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("schedule: write: %w", err)
	}
	return nil
}

// Load reads a schedule from a JSON file at the given path.
func Load(path string) (*Schedule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Schedule{}, nil
		}
		return nil, fmt.Errorf("schedule: read: %w", err)
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("schedule: unmarshal: %w", err)
	}
	return &s, nil
}

// DueEntries returns all entries that are due to run at the given time.
func DueEntries(s *Schedule, now time.Time) []Entry {
	var due []Entry
	for _, e := range s.Entries {
		if e.IsDue(now) {
			due = append(due, e)
		}
	}
	return due
}
