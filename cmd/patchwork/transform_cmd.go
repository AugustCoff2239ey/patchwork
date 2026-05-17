package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/patchwork/internal/snapshot"
	"github.com/yourorg/patchwork/internal/transform"
)

func runTransform(args []string) error {
	fs := flag.NewFlagSet("transform", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config JSON file")
	env := fs.String("env", "", "environment label")
	opsPath := fs.String("ops", "", "path to transform ops JSON file")
	outPath := fs.String("out", "", "path to write transformed snapshot (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("transform: --config is required")
	}
	if *env == "" {
		return fmt.Errorf("transform: --env is required")
	}
	if *opsPath == "" {
		return fmt.Errorf("transform: --ops is required")
	}

	snap, err := snapshot.Capture(*configPath, *env)
	if err != nil {
		return err
	}

	opsData, err := os.ReadFile(*opsPath)
	if err != nil {
		return fmt.Errorf("transform: read ops file: %w", err)
	}
	var ops []transform.Op
	if err := json.Unmarshal(opsData, &ops); err != nil {
		return fmt.Errorf("transform: parse ops: %w", err)
	}

	result, err := transform.Apply(snap, ops)
	if err != nil {
		return err
	}

	fmt.Print(transform.Render(result))

	if *outPath != "" {
		if err := snapshot.Save(*outPath, result.Snapshot); err != nil {
			return err
		}
		fmt.Printf("Saved transformed snapshot to %s\n", *outPath)
	}
	return nil
}
