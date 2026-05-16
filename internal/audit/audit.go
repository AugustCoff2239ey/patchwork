package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// EventKind describes the type of audit event.
type EventKind string

const (
	EventCapture  EventKind = "capture"
	EventDiff     EventKind = "diff"
	EventBaseline EventKind = "baseline"
	EventRollback EventKind = "rollback"
	EventPrune    EventKind = "prune"
	EventExport   EventKind = "export"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp   time.Time `json:"timestamp"`
	Kind        EventKind `json:"kind"`
	Environment string    `json:"environment"`
	Message     string    `json:"message"`
	User        string    `json:"user,omitempty"`
}

// Log holds an ordered list of audit events.
type Log struct {
	Events []Event `json:"events"`
}

// Append adds a new event to the log.
func Append(log *Log, kind EventKind, env, message string) {
	user := os.Getenv("USER")
	log.Events = append(log.Events, Event{
		Timestamp:   time.Now().UTC(),
		Kind:        kind,
		Environment: env,
		Message:     message,
		User:        user,
	})
}

// Filter returns events matching the given kind and/or environment.
// Pass empty strings to skip filtering on that field.
func Filter(log *Log, kind EventKind, env string) []Event {
	var out []Event
	for _, e := range log.Events {
		if kind != "" && e.Kind != kind {
			continue
		}
		if env != "" && e.Environment != env {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Save writes the audit log to a JSON file.
func Save(path string, log *Log) error {
	sort.Slice(log.Events, func(i, j int) bool {
		return log.Events[i].Timestamp.Before(log.Events[j].Timestamp)
	})
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads an audit log from a JSON file.
func Load(path string) (*Log, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Log{}, nil
		}
		return nil, fmt.Errorf("audit: read: %w", err)
	}
	var log Log
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("audit: unmarshal: %w", err)
	}
	return &log, nil
}

// Render returns a human-readable string of audit events.
func Render(events []Event) string {
	if len(events) == 0 {
		return "No audit events found.\n"
	}
	out := "Audit Log:\n"
	for _, e := range events {
		user := e.User
		if user == "" {
			user = "unknown"
		}
		out += fmt.Sprintf("  [%s] %-10s env=%-12s user=%-10s %s\n",
			e.Timestamp.Format(time.RFC3339), e.Kind, e.Environment, user, e.Message)
	}
	return out
}
