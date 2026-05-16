package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/patchwork/internal/snapshot"
	"github.com/patchwork/internal/validate"
)

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config file to validate")
	rulesPath := fs.String("rules", "", "path to JSON rules file")
	env := fs.String("env", "default", "environment label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("--config flag is required")
	}
	if *rulesPath == "" {
		return fmt.Errorf("--rules flag is required")
	}

	snap, err := snapshot.Capture(*configPath, *env)
	if err != nil {
		return fmt.Errorf("capturing snapshot: %w", err)
	}

	rules, err := loadValidateRules(*rulesPath)
	if err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	report := validate.Run(snap, rules)
	fmt.Print(validate.Render(report))

	if report.Failed > 0 {
		return fmt.Errorf("%d validation rule(s) failed", report.Failed)
	}
	return nil
}

func loadValidateRules(path string) ([]validate.Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rules []validate.Rule
	if err := json.NewDecoder(f).Decode(&rules); err != nil {
		return nil, fmt.Errorf("decoding rules JSON: %w", err)
	}
	return rules, nil
}
