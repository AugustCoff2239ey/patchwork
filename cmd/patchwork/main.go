package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/snapshot"
)

const usage = `patchwork — lightweight config drift tracker

Usage:
  patchwork capture <config-file> <output-snapshot>
  patchwork diff    <snapshot-a>  <snapshot-b>
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "capture":
		runCapture(args[1:])
	case "diff":
		runDiff(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n", args[0])
		flag.Usage()
		os.Exit(1)
	}
}

func runCapture(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: patchwork capture <config-file> <output-snapshot>")
		os.Exit(1)
	}
	configPath, outputPath := args[0], args[1]

	snap, err := snapshot.Capture(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := snapshot.Save(snap, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Snapshot saved to %s (%d entries)\n", outputPath, len(snap.Entries))
}

func runDiff(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: patchwork diff <snapshot-a> <snapshot-b>")
		os.Exit(1)
	}

	from, err := snapshot.Load(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading snapshot a: %v\n", err)
		os.Exit(1)
	}

	to, err := snapshot.Load(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading snapshot b: %v\n", err)
		os.Exit(1)
	}

	result := diff.Compare(from, to)
	fmt.Print(diff.Format(result))
}
