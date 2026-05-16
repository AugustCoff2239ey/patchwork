package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/patch"
	"github.com/patchwork/internal/snapshot"
)

func runPatch(args []string) error {
	fs := flag.NewFlagSet("patch", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file to patch")
	env := fs.String("env", "", "environment name")
	opsFile := fs.String("ops", "", "path to JSON file containing patch operations")
	dryRun := fs.Bool("dry-run", false, "print result without writing")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("patch: --config flag is required")
	}
	if *env == "" {
		return fmt.Errorf("patch: --env flag is required")
	}
	if *opsFile == "" {
		return fmt.Errorf("patch: --ops flag is required")
	}

	snap, err := snapshot.Capture(*configPath, *env)
	if err != nil {
		return fmt.Errorf("patch: capture failed: %w", err)
	}

	opsData, err := os.ReadFile(*opsFile)
	if err != nil {
		return fmt.Errorf("patch: reading ops file: %w", err)
	}

	var ops []patch.Op
	if err := json.Unmarshal(opsData, &ops); err != nil {
		return fmt.Errorf("patch: parsing ops: %w", err)
	}

	plan := patch.Plan{
		Environment: *env,
		Ops:         ops,
	}

	result, err := patch.Apply(plan, snap)
	if err != nil {
		return fmt.Errorf("patch: apply failed: %w", err)
	}

	fmt.Print(patch.Render(result))

	if *dryRun {
		fmt.Println("(dry-run: no changes written)")
		return nil
	}

	if err := snapshot.Save(result.Snap, *configPath+".patched"); err != nil {
		return fmt.Errorf("patch: saving result: %w", err)
	}
	fmt.Printf("Patched snapshot written to %s.patched\n", *configPath)
	return nil
}
