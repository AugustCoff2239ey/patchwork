package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/index"
)

func runIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	historyPath := fs.String("history", "patchwork-history.json", "Path to history file")
	query := fs.String("query", "", "Search query (key or value substring)")
	env := fs.String("env", "", "Filter by environment")
	save := fs.String("save", "", "Save index to file at this path")
	load := fs.String("load", "", "Load index from file instead of building from history")
	_ = fs.Parse(args)

	var idx index.Index

	if *load != "" {
		loaded, err := index.Load(*load)
		if err != nil {
			fmt.Fprintf(os.Stderr, "index: load: %v\n", err)
			os.Exit(1)
		}
		idx = loaded
	} else {
		log, err := history.LoadLog(*historyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "index: load history: %v\n", err)
			os.Exit(1)
		}
		if len(log) == 0 {
			fmt.Println("No history entries found.")
			return
		}
		idx = index.Build(log)
	}

	if *save != "" {
		if err := index.Save(idx, *save); err != nil {
			fmt.Fprintf(os.Stderr, "index: save: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Index saved to %s (%d entries)\n", *save, len(idx.Entries))
		return
	}

	if *query == "" {
		fmt.Printf("Index contains %d entries across all environments.\n", len(idx.Entries))
		return
	}

	results := index.Search(idx, *query, *env)
	fmt.Print(index.Render(results))
}
