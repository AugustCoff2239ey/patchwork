package watch

import (
	"fmt"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/snapshot"
)

// WatchConfig holds configuration for a watch session.
type WatchConfig struct {
	FilePath    string
	Environment string
	Interval    time.Duration
	MaxDrifts   int
}

// DriftEvent represents a detected configuration drift during a watch session.
type DriftEvent struct {
	DetectedAt time.Time
	Changes    []diff.Change
}

// Watcher polls a config file and emits drift events when changes are detected.
type Watcher struct {
	cfg      WatchConfig
	last     *snapshot.Snapshot
	events   []DriftEvent
}

// New creates a new Watcher for the given config.
func New(cfg WatchConfig) *Watcher {
	return &Watcher{cfg: cfg}
}

// Poll captures the current state of the file and compares it to the last
// known snapshot. Returns a DriftEvent if changes are detected, nil otherwise.
func (w *Watcher) Poll() (*DriftEvent, error) {
	current, err := snapshot.Capture(w.cfg.FilePath, w.cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("watch: capture failed: %w", err)
	}

	if w.last == nil {
		w.last = current
		return nil, nil
	}

	changes := diff.Compare(*w.last, *current)
	if len(changes) == 0 {
		return nil, nil
	}

	event := &DriftEvent{
		DetectedAt: time.Now().UTC(),
		Changes:    changes,
	}
	w.last = current
	w.events = append(w.events, *event)
	return event, nil
}

// Events returns all drift events collected during this watch session.
func (w *Watcher) Events() []DriftEvent {
	return w.events
}

// DriftCount returns the total number of drift events detected.
func (w *Watcher) DriftCount() int {
	return len(w.events)
}
