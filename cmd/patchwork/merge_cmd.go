package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/merge"
	"github.com/patchwork/internal/snapshot"
)

// runMerge merges two snapshot files and writes the result to an output path.
// Usage: patchwork merge --base <file> --incoming <file> --out <file> [--strategy ours|theirs|error]
func runMerge(args []string) error {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	baseFlag := fs.String("base", "", "path to base snapshot file (required)")
	incomingFlag := fs.String("incoming", "", "path to incoming snapshot file (required)")
	outFlag := fs.String("out", "", "path to write merged snapshot (required)")
	strategyFlag := fs.String("strategy", "ours", "conflict resolution strategy: ours, theirs, error")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *baseFlag == "" || *incomingFlag == "" || *outFlag == "" {
		return fmt.Errorf("--base, --incoming, and --out are required")
	}

	base, err := snapshot.Load(*baseFlag)
	if err != nil {
		return fmt.Errorf("loading base snapshot: %w", err)
	}
	incoming, err := snapshot.Load(*incomingFlag)
	if err != nil {
		return fmt.Errorf("loading incoming snapshot: %w", err)
	}

	strategy := merge.Strategy(*strategyFlag)
	result, err := merge.Apply(base, incoming, strategy)
	if err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	if err := snapshot.Save(result.Snapshot, *outFlag); err != nil {
		return fmt.Errorf("saving merged snapshot: %w", err)
	}

	fmt.Fprint(os.Stdout, merge.Render(result))
	return nil
}
