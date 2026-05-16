package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/drift"
	"github.com/patchwork/internal/history"
)

func runDrift(args []string) {
	fs := flag.NewFlagSet("drift", flag.ExitOnError)
	historyFile := fs.String("history", "patchwork_history.json", "Path to history log file")
	days := fs.Int("days", 7, "Number of days to analyze")
	fs.Parse(args)

	if *days <= 0 {
		fmt.Fprintln(os.Stderr, "error: --days must be a positive integer")
		os.Exit(1)
	}

	log, err := history.LoadLog(*historyFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error loading history: %v\n", err)
		os.Exit(1)
	}

	report := drift.Analyze(log, *days)
	fmt.Print(drift.Render(report))
}
