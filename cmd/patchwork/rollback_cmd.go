package main

import (
	"fmt"
	"os"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/rollback"
)

// runRollback loads the history log, prepares a rollback plan for the given
// environment, and applies it, writing the restored snapshot to destPath.
func runRollback(env, historyPath, destPath string) error {
	if env == "" {
		return fmt.Errorf("rollback: environment must be specified")
	}

	log, err := history.LoadLog(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("rollback: no history file found at %s", historyPath)
		}
		return fmt.Errorf("rollback: load history: %w", err)
	}

	plan, err := rollback.Prepare(env, log)
	if err != nil {
		return fmt.Errorf("rollback: prepare: %w", err)
	}

	fmt.Printf("Rolling back environment %q\n", plan.Environment)
	fmt.Printf("  From: %s (%s)\n", plan.FromEntry.SnapshotPath, plan.FromEntry.CapturedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  To:   %s (%s)\n", plan.ToEntry.SnapshotPath, plan.ToEntry.CapturedAt.Format("2006-01-02 15:04:05"))

	result, err := rollback.Apply(plan, destPath)
	if err != nil {
		return fmt.Errorf("rollback: apply: %w", err)
	}

	fmt.Println(result.Message)
	fmt.Printf("Snapshot written to %s\n", destPath)
	return nil
}
