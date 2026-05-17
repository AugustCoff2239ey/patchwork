package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/scope"
	"github.com/patchwork/internal/snapshot"
)

func runScope(args []string) error {
	fs := flag.NewFlagSet("scope", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file to snapshot")
	env := fs.String("env", "", "environment name")
	scopeName := fs.String("name", "", "scope name")
	keysFlag := fs.String("keys", "", "comma-separated list of keys to include in scope")
	scopeFile := fs.String("scope-file", "", "path to JSON scope definition file (overrides other flags)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	snap, err := snapshot.Capture(*configPath, *env)
	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}

	var sc scope.Scope
	if *scopeFile != "" {
		data, err := os.ReadFile(*scopeFile)
		if err != nil {
			return fmt.Errorf("reading scope file: %w", err)
		}
		if err := json.Unmarshal(data, &sc); err != nil {
			return fmt.Errorf("parsing scope file: %w", err)
		}
	} else {
		if *scopeName == "" {
			return fmt.Errorf("--name is required")
		}
		if *keysFlag == "" {
			return fmt.Errorf("--keys is required")
		}
		sc = scope.Scope{
			Name:        *scopeName,
			Environment: *env,
			Keys:        splitComma(*keysFlag),
		}
	}

	result, err := scope.Apply(snap, sc)
	if err != nil {
		return fmt.Errorf("scope apply failed: %w", err)
	}

	fmt.Print(scope.Render(result))
	return nil
}

func splitComma(s string) []string {
	var out []string
	for _, part := range splitString(s, ',') {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func splitString(s string, sep rune) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}
