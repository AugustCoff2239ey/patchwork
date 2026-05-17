package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/patchwork/internal/snapshot"
)

func runDelta(args []string) {
	fs := flag.NewFlagSet("delta", flag.ExitOnError)
	before := fs.String("before", "", "path to the older snapshot file")
	after := fs.String("after", "", "path to the newer snapshot file")
	fs.Parse(args)

	if *before == "" {
		fmt.Fprintln(os.Stderr, "error: --before flag is required")
		os.Exit(1)
	}
	if *after == "" {
		fmt.Fprintln(os.Stderr, "error: --after flag is required")
		os.Exit(1)
	}

	snBefore, err := snapshot.Load(*before)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading before snapshot: %v\n", err)
		os.Exit(1)
	}

	snAfter, err := snapshot.Load(*after)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading after snapshot: %v\n", err)
		os.Exit(1)
	}

	d, err := snapshot.Delta(snBefore, snAfter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing delta: %v\n", err)
		os.Exit(1)
	}

	snapshot.RenderDelta(os.Stdout, d)
}
