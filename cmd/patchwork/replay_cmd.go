package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/replay"
	"github.com/patchwork/internal/snapshot"
)

// runReplay loads history for the given environment and renders a chronological
// replay of all captured snapshots, showing configuration state at each point.
func runReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	env := fs.String("env", "", "environment to replay (required)")
	historyPath := fs.String("history", ".patchwork/history.json", "path to history log")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *env == "" {
		return fmt.Errorf("replay: --env flag is required")
	}

	log, err := history.LoadLog(*historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("replay: history file not found: %s", *historyPath)
		}
		return fmt.Errorf("replay: loading history: %w", err)
	}

	loadFn := func(path string) (snapshot.Snapshot, error) {
		return snapshot.Load(path)
	}

	result, err := replay.Build(log, *env, loadFn)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}

	fmt.Print(replay.Render(result))
	return nil
}
