package main

import (
	"fmt"
	"os"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/policy"
	"github.com/patchwork/internal/snapshot"
)

func runPolicy(policyPath, historyPath, env string) error {
	if policyPath == "" {
		return fmt.Errorf("policy: --policy flag is required")
	}
	if env == "" {
		return fmt.Errorf("policy: --env flag is required")
	}

	pf, err := policy.Load(policyPath)
	if err != nil {
		return fmt.Errorf("policy: load rules: %w", err)
	}

	log, err := history.LoadLog(historyPath)
	if err != nil {
		return fmt.Errorf("policy: load history: %w", err)
	}

	var envEntries []history.Entry
	for _, e := range log.Entries {
		if e.Environment == env {
			envEntries = append(envEntries, e)
		}
	}

	if len(envEntries) < 2 {
		fmt.Println("policy: not enough history entries to evaluate")
		return nil
	}

	latest := envEntries[len(envEntries)-1]
	previous := envEntries[len(envEntries)-2]

	snap1, err := snapshot.Load(previous.SnapshotPath)
	if err != nil {
		return fmt.Errorf("policy: load previous snapshot: %w", err)
	}
	snap2, err := snapshot.Load(latest.SnapshotPath)
	if err != nil {
		return fmt.Errorf("policy: load latest snapshot: %w", err)
	}

	changes := diff.Compare(snap1, snap2)
	violations := policy.Evaluate(pf, env, changes)
	fmt.Print(policy.Render(violations))

	if len(violations) > 0 {
		os.Exit(1)
	}
	return nil
}
