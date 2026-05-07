package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/patchwork/internal/baseline"
	"github.com/user/patchwork/internal/diff"
	"github.com/user/patchwork/internal/snapshot"
)

func baselinePath(env string) string {
	return filepath.Join(".patchwork", "baselines", env+".json")
}

// runBaseline pins the current snapshot of configFile as the baseline for env.
func runBaseline(env, configFile string) error {
	snap, err := snapshot.Capture(configFile)
	if err != nil {
		return fmt.Errorf("capture snapshot: %w", err)
	}

	b := baseline.Pin(env, snap)
	path := baselinePath(env)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create baseline dir: %w", err)
	}

	if err := baseline.Save(path, b); err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}

	fmt.Printf("Baseline pinned for environment %q at %s\n", env, b.PinnedAt.Format("2006-01-02 15:04:05"))
	return nil
}

// runBaselineDiff compares the current configFile against the pinned baseline for env.
func runBaselineDiff(env, configFile string) error {
	path := baselinePath(env)
	b, err := baseline.Load(path)
	if err != nil {
		return fmt.Errorf("load baseline: %w", err)
	}

	current, err := snapshot.Capture(configFile)
	if err != nil {
		return fmt.Errorf("capture snapshot: %w", err)
	}

	changes := diff.Compare(b.Snapshot, current)
	if len(changes) == 0 {
		fmt.Printf("No drift detected from baseline (pinned %s)\n", b.PinnedAt.Format("2006-01-02 15:04:05"))
		return nil
	}

	fmt.Printf("Drift from baseline (pinned %s):\n\n", b.PinnedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(diff.Format(changes))
	return nil
}
