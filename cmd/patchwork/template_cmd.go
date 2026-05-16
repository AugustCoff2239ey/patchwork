package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/snapshot"
	"github.com/patchwork/internal/template"
)

// runTemplate applies a JSON template file to a captured snapshot and prints
// the result. Usage:
//
//	patchwork template --tmpl <file> --config <file> --env <env> [--overwrite]
func runTemplate(args []string) error {
	fs := flag.NewFlagSet("template", flag.ContinueOnError)
	tmplPath := fs.String("tmpl", "", "path to template JSON file")
	configPath := fs.String("config", "", "path to config file to snapshot")
	env := fs.String("env", "", "environment label")
	overwrite := fs.Bool("overwrite", false, "overwrite existing keys")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tmplPath == "" {
		return fmt.Errorf("--tmpl flag is required")
	}
	if *configPath == "" {
		return fmt.Errorf("--config flag is required")
	}
	if *env == "" {
		return fmt.Errorf("--env flag is required")
	}

	// Load template
	tmplData, err := os.ReadFile(*tmplPath)
	if err != nil {
		return fmt.Errorf("reading template: %w", err)
	}
	var tmpl template.Template
	if err := json.Unmarshal(tmplData, &tmpl); err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// Capture snapshot
	snap, err := snapshot.Capture(*configPath, *env)
	if err != nil {
		return fmt.Errorf("capturing snapshot: %w", err)
	}

	// Apply template
	_, result, err := template.Apply(tmpl, snap, *overwrite)
	if err != nil {
		return fmt.Errorf("applying template: %w", err)
	}

	fmt.Print(template.Render(result))
	return nil
}
