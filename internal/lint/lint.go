package lint

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Rule defines a linting rule applied to a snapshot.
type Rule struct {
	Name    string
	Message string
	Check   func(snap snapshot.Snapshot) bool
}

// Result holds the outcome of a single rule evaluation.
type Result struct {
	Rule    string
	Passed  bool
	Message string
}

// Report aggregates all lint results for a snapshot.
type Report struct {
	Environment string
	Results     []Result
	PassCount   int
	FailCount   int
}

// DefaultRules returns the built-in linting rules.
func DefaultRules() []Rule {
	return []Rule{
		{
			Name:    "no-empty-values",
			Message: "snapshot contains keys with empty values",
			Check: func(snap snapshot.Snapshot) bool {
				for _, v := range snap.Data {
					if strings.TrimSpace(v) == "" {
						return false
					}
				}
				return true
			},
		},
		{
			Name:    "no-duplicate-keys",
			Message: "snapshot data map should not contain duplicate keys (always passes in Go maps)",
			Check: func(snap snapshot.Snapshot) bool {
				return true
			},
		},
		{
			Name:    "has-entries",
			Message: "snapshot contains no configuration keys",
			Check: func(snap snapshot.Snapshot) bool {
				return len(snap.Data) > 0
			},
		},
	}
}

// Run evaluates all rules against the given snapshot and returns a Report.
func Run(snap snapshot.Snapshot, rules []Rule) Report {
	report := Report{Environment: snap.Environment}
	for _, rule := range rules {
		passed := rule.Check(snap)
		msg := ""
		if !passed {
			msg = rule.Message
			report.FailCount++
		} else {
			report.PassCount++
		}
		report.Results = append(report.Results, Result{
			Rule:    rule.Name,
			Passed:  passed,
			Message: msg,
		})
	}
	return report
}

// Render formats the lint report as a human-readable string.
func Render(r Report) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Lint Report [%s]\n", r.Environment))
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	for _, res := range r.Results {
		status := "PASS"
		if !res.Passed {
			status = "FAIL"
		}
		line := fmt.Sprintf("  [%s] %s", status, res.Rule)
		if res.Message != "" {
			line += fmt.Sprintf(": %s", res.Message)
		}
		sb.WriteString(line + "\n")
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d passed, %d failed\n", r.PassCount, r.FailCount))
	return sb.String()
}
