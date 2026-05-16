package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/normalize"
	"github.com/patchwork/internal/snapshot"
)

func runNormalize(args []string) error {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	config := fs.String("config", "", "path to config file")
	env := fs.String("env", "", "environment name")
	trimSpace := fs.Bool("trim", true, "trim whitespace from keys and values")
	lowerKeys := fs.Bool("lowercase-keys", false, "convert keys to lowercase")
	removeEmpty := fs.Bool("remove-empty", false, "remove keys with empty values")
	dryRun := fs.Bool("dry-run", false, "print changes without writing output")
	output := fs.String("output", "", "write normalized snapshot to this file (JSON)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *config == "" {
		return fmt.Errorf("normalize: --config is required")
	}
	if *env == "" {
		return fmt.Errorf("normalize: --env is required")
	}

	snap, err := snapshot.Capture(*config, *env)
	if err != nil {
		return fmt.Errorf("normalize: capture failed: %w", err)
	}

	opts := normalize.Options{
		TrimSpace:    *trimSpace,
		LowercaseKeys: *lowerKeys,
		RemoveEmpty:  *removeEmpty,
	}

	result, err := normalize.Apply(snap, opts)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}

	fmt.Print(normalize.Render(result))

	if *dryRun {
		return nil
	}

	dest := *output
	if dest == "" {
		dest = *config + ".normalized.json"
	}

	if err := snapshot.Save(result.Snapshot, dest); err != nil {
		return fmt.Errorf("normalize: save failed: %w", err)
	}

	fmt.Fprintf(os.Stdout, "normalize: written to %s\n", dest)
	return nil
}
