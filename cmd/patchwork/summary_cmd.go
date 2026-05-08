package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/summary"
)

func runSummary(args []string) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	env := fs.String("env", "", "filter by environment (optional)")
	logPath := fs.String("log", "patchwork.log.json", "path to history log file")
	fs.Parse(args)

	log, err := history.LoadLog(*logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "no history log found")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error loading log: %v\n", err)
		os.Exit(1)
	}

	if len(log.Entries) == 0 {
		fmt.Println("no history entries found")
		return
	}

	stats := summary.Build(log.Entries, *env)
	fmt.Print(summary.Render(stats))
}
