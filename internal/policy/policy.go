package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/patchwork/internal/diff"
)

// Rule defines a single policy rule applied to config changes.
type Rule struct {
	Name        string   `json:"name"`
	Environments []string `json:"environments,omitempty"`
	ForbidKeys  []string `json:"forbid_keys,omitempty"`
	RequireKeys []string `json:"require_keys,omitempty"`
	MaxChanges  int      `json:"max_changes,omitempty"`
}

// Violation describes a rule that was not satisfied.
type Violation struct {
	Rule    string
	Message string
}

// PolicyFile holds a list of rules loaded from disk.
type PolicyFile struct {
	Rules []Rule `json:"rules"`
}

// Load reads a policy file from the given path.
func Load(path string) (PolicyFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PolicyFile{}, fmt.Errorf("policy: read file: %w", err)
	}
	var pf PolicyFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return PolicyFile{}, fmt.Errorf("policy: parse: %w", err)
	}
	return pf, nil
}

// Evaluate checks a set of changes against all rules and returns violations.
func Evaluate(pf PolicyFile, env string, changes []diff.Change) []Violation {
	var violations []Violation
	for _, rule := range pf.Rules {
		if !appliesToEnv(rule, env) {
			continue
		}
		if rule.MaxChanges > 0 && len(changes) > rule.MaxChanges {
			violations = append(violations, Violation{
				Rule:    rule.Name,
				Message: fmt.Sprintf("change count %d exceeds max %d", len(changes), rule.MaxChanges),
			})
		}
		for _, key := range rule.ForbidKeys {
			for _, c := range changes {
				if strings.EqualFold(c.Key, key) {
					violations = append(violations, Violation{
						Rule:    rule.Name,
						Message: fmt.Sprintf("forbidden key changed: %s", c.Key),
					})
				}
			}
		}
		for _, key := range rule.RequireKeys {
			found := false
			for _, c := range changes {
				if strings.EqualFold(c.Key, key) {
					found = true
					break
				}
			}
			if !found {
				violations = append(violations, Violation{
					Rule:    rule.Name,
					Message: fmt.Sprintf("required key not changed: %s", key),
				})
			}
		}
	}
	return violations
}

// Render formats violations for display.
func Render(violations []Violation) string {
	if len(violations) == 0 {
		return "policy: all rules passed\n"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("policy: %d violation(s) found\n", len(violations)))
	for _, v := range violations {
		sb.WriteString(fmt.Sprintf("  [%s] %s\n", v.Rule, v.Message))
	}
	return sb.String()
}

func appliesToEnv(rule Rule, env string) bool {
	if len(rule.Environments) == 0 {
		return true
	}
	for _, e := range rule.Environments {
		if strings.EqualFold(e, env) {
			return true
		}
	}
	return false
}
