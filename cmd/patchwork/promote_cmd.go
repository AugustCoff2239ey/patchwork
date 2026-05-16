package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/patchwork/internal/promote"
	"github.com/patchwork/internal/snapshot"
)

func runPromote(args []string) error {
	fs := flag.NewFlagSet("promote", flag.ContinueOnError)
	srcFile := fs.String("src", "", "path to source snapshot file")
	dstFile := fs.String("dst", "", "path to target snapshot file (use - for empty)")
	targetEnv := fs.String("env", "", "target environment name")
	onlyKeys := fs.String("keys", "", "comma-separated list of keys to promote")
	overwrite := fs.Bool("overwrite", false, "overwrite existing keys in target")
	outFile := fs.String("out", "", "write resulting snapshot to this file")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *srcFile == "" {
		return fmt.Errorf("--src is required")
	}
	if *targetEnv == "" {
		return fmt.Errorf("--env is required")
	}

	src, err := snapshot.Load(*srcFile)
	if err != nil {
		return fmt.Errorf("load source: %w", err)
	}

	var dst snapshot.Snapshot
	if *dstFile == "" || *dstFile == "-" {
		dst = snapshot.Snapshot{
			Environment: *targetEnv,
			Timestamp:   time.Now(),
			Data:        map[string]string{},
		}
	} else {
		dst, err = snapshot.Load(*dstFile)
		if err != nil {
			return fmt.Errorf("load target: %w", err)
		}
		dst.Environment = *targetEnv
	}

	opts := promote.Options{Overwrite: *overwrite}
	if *onlyKeys != "" {
		opts.OnlyKeys = strings.Split(*onlyKeys, ",")
	}

	result, plan, err := promote.Apply(src, dst, opts)
	if err != nil {
		return err
	}

	fmt.Print(promote.Render(plan))

	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(result); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Snapshot written to %s\n", *outFile)
	}

	return nil
}
