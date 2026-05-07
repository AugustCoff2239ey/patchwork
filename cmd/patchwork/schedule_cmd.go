package main

import (
	"fmt"
	"os"
	"time"

	"github.com/user/patchwork/internal/schedule"
	"github.com/user/patchwork/internal/snapshot"
)

const defaultSchedulePath = ".patchwork/schedule.json"

// runScheduleAdd registers a new scheduled capture entry.
func runScheduleAdd(env, configPath string, interval time.Duration) error {
	s, err := schedule.Load(defaultSchedulePath)
	if err != nil {
		return fmt.Errorf("load schedule: %w", err)
	}

	for _, e := range s.Entries {
		if e.Environment == env && e.ConfigPath == configPath {
			return fmt.Errorf("schedule entry for env=%q path=%q already exists", env, configPath)
		}
	}

	s.Entries = append(s.Entries, schedule.Entry{
		Environment: env,
		ConfigPath:  configPath,
		Interval:    interval,
	})

	if err := schedule.Save(defaultSchedulePath, s); err != nil {
		return fmt.Errorf("save schedule: %w", err)
	}

	fmt.Printf("scheduled: env=%s path=%s interval=%s\n", env, configPath, interval)
	return nil
}

// runScheduleRun executes all due captures based on the current schedule.
func runScheduleRun() error {
	s, err := schedule.Load(defaultSchedulePath)
	if err != nil {
		return fmt.Errorf("load schedule: %w", err)
	}

	now := time.Now()
	due := schedule.DueEntries(s, now)

	if len(due) == 0 {
		fmt.Println("no entries due for capture")
		return nil
	}

	for i := range s.Entries {
		e := &s.Entries[i]
		if !e.IsDue(now) {
			continue
		}

		snap, err := snapshot.Capture(e.ConfigPath, e.Environment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "capture failed for %s: %v\n", e.Environment, err)
			continue
		}

		snapPath := fmt.Sprintf(".patchwork/snapshots/%s-%d.json", e.Environment, now.Unix())
		if err := snapshot.Save(snapPath, snap); err != nil {
			fmt.Fprintf(os.Stderr, "save snapshot failed for %s: %v\n", e.Environment, err)
			continue
		}

		e.MarkRun(now)
		fmt.Printf("captured: env=%s -> %s\n", e.Environment, snapPath)
	}

	return schedule.Save(defaultSchedulePath, s)
}
