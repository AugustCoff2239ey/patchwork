package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/archive"
	"github.com/patchwork/internal/history"
)

func runArchive(args []string) error {
	fs := flag.NewFlagSet("archive", flag.ContinueOnError)
	historyPath := fs.String("history", "patchwork-history.json", "path to history log")
	archivePath := fs.String("out", "patchwork-archive.json", "path to write archive")
	env := fs.String("env", "", "filter by environment (optional)")
	listOnly := fs.Bool("list", false, "list archive entries without writing")

	if err := fs.Parse(args); err != nil {
		return err
	}

	log, err := history.LoadLog(*historyPath)
	if err != nil {
		return fmt.Errorf("archive: load history: %w", err)
	}

	if len(log.Entries) == 0 {
		fmt.Fprintln(os.Stdout, "archive: no history entries found")
		return nil
	}

	a := archive.Build(log, *env)

	if *listOnly {
		fmt.Print(archive.Render(a))
		return nil
	}

	// Merge with existing archive if present.
	existing, err := archive.Load(*archivePath)
	if err != nil {
		return fmt.Errorf("archive: load existing: %w", err)
	}

	seen := map[string]bool{}
	for _, e := range existing.Entries {
		seen[e.ID] = true
	}
	for _, e := range a.Entries {
		if !seen[e.ID] {
			existing.Entries = append(existing.Entries, e)
		}
	}

	if err := archive.Save(*archivePath, existing); err != nil {
		return fmt.Errorf("archive: save: %w", err)
	}

	fmt.Printf("archive: wrote %d entries to %s\n", len(existing.Entries), *archivePath)
	return nil
}
