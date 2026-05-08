package main

import (
	"fmt"
	"os"

	"github.com/patchwork/internal/lint"
	"github.com/patchwork/internal/snapshot"
)

// runLint captures a snapshot from the given config file and runs lint rules against it.
func runLint(configPath, environment string, strict bool) error {
	if configPath == "" {
		return fmt.Errorf("config path is required")
	}
	if environment == "" {
		return fmt.Errorf("environment is required")
	}

	snap, err := snapshot.Capture(configPath, environment)
	if err != nil {
		return fmt.Errorf("failed to capture snapshot: %w", err)
	}

	rules := lint.DefaultRules()
	report := lint.Run(snap, rules)
	output := lint.Render(report)
	fmt.Print(output)

	if strict && report.FailCount > 0 {
		return fmt.Errorf("lint failed: %d rule(s) did not pass", report.FailCount)
	}
	return nil
}

func init() {
	// Register lint subcommand in main arg parsing.
	_ = os.Args // referenced to satisfy import
}
