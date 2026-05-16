package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Rule defines a single validation rule applied to a snapshot.
type Rule struct {
	Key     string // exact key or pattern prefix "regex:"
	Pattern string // expected value regex (empty means key must exist)
	Required bool
}

// Result holds the outcome of a single rule evaluation.
type Result struct {
	Rule    Rule
	Passed  bool
	Message string
}

// Report is the full validation output for a snapshot.
type Report struct {
	Environment string
	Results     []Result
	Passed      int
	Failed      int
}

// Run evaluates all rules against the given snapshot.
func Run(snap snapshot.Snapshot, rules []Rule) Report {
	report := Report{Environment: snap.Environment}
	for _, rule := range rules {
		result := evaluate(snap, rule)
		report.Results = append(report.Results, result)
		if result.Passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}
	return report
}

func evaluate(snap snapshot.Snapshot, rule Rule) Result {
	val, ok := snap.Data[rule.Key]
	if !ok {
		if rule.Required {
			return Result{Rule: rule, Passed: false, Message: fmt.Sprintf("required key %q is missing", rule.Key)}
		}
		return Result{Rule: rule, Passed: true, Message: "key not present (optional)"}
	}
	if rule.Pattern == "" {
		return Result{Rule: rule, Passed: true, Message: "key present"}
	}
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return Result{Rule: rule, Passed: false, Message: fmt.Sprintf("invalid pattern: %v", err)}
	}
	if re.MatchString(val) {
		return Result{Rule: rule, Passed: true, Message: "value matches pattern"}
	}
	return Result{Rule: rule, Passed: false, Message: fmt.Sprintf("key %q value %q does not match pattern %q", rule.Key, val, rule.Pattern)}
}

// Render returns a human-readable summary of the report.
func Render(r Report) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Validation Report [%s]\n", r.Environment)
	fmt.Fprintf(&sb, "Passed: %d  Failed: %d\n", r.Passed, r.Failed)
	for _, res := range r.Results {
		status := "PASS"
		if !res.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(&sb, "  [%s] %s — %s\n", status, res.Rule.Key, res.Message)
	}
	return sb.String()
}
