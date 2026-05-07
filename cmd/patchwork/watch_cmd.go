package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/watch"
)

// runWatch polls a config file at a given interval and prints drift events
// until interrupted or the max drift count is reached.
func runWatch(filePath, env string, interval time.Duration, maxDrifts int) error {
	if filePath == "" {
		return fmt.Errorf("watch: file path is required")
	}
	if env == "" {
		env = "default"
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}

	w := watch.New(watch.WatchConfig{
		FilePath:    filePath,
		Environment: env,
		Interval:    interval,
		MaxDrifts:   maxDrifts,
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Fprintf(os.Stdout, "Watching %s (env=%s, interval=%s)\n", filePath, env, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Prime the watcher with an initial snapshot.
	if _, err := w.Poll(); err != nil {
		return fmt.Errorf("watch: initial poll failed: %w", err)
	}

	for {
		select {
		case <-sigs:
			fmt.Fprintf(os.Stdout, "\nWatch stopped. Total drift events: %d\n", w.DriftCount())
			return nil
		case <-ticker.C:
			event, err := w.Poll()
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch: poll error: %v\n", err)
				continue
			}
			if event == nil {
				continue
			}
			fmt.Fprintf(os.Stdout, "[%s] Drift detected (%d change(s)):\n",
				event.DetectedAt.Format(time.RFC3339), len(event.Changes))
			fmt.Fprintln(os.Stdout, diff.Format(event.Changes))

			if maxDrifts > 0 && w.DriftCount() >= maxDrifts {
				fmt.Fprintf(os.Stdout, "Max drift count (%d) reached. Stopping.\n", maxDrifts)
				return nil
			}
		}
	}
}
