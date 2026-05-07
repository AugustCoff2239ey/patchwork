package alert

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/diff"
)

// Severity represents the alert level for a detected change.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Alert represents a single drift alert for a changed key.
type Alert struct {
	Key      string
	Severity Severity
	Message  string
}

// Rules maps key prefixes or exact keys to a severity level.
type Rules map[string]Severity

// DefaultRules provides a basic set of severity rules.
var DefaultRules = Rules{
	"password": SeverityCritical,
	"secret":   SeverityCritical,
	"token":    SeverityCritical,
	"host":     SeverityWarning,
	"port":     SeverityWarning,
	"endpoint": SeverityWarning,
}

// Evaluate inspects a slice of diff.Changes and produces alerts based on the provided rules.
func Evaluate(changes []diff.Change, rules Rules) []Alert {
	var alerts []Alert
	for _, c := range changes {
		sev := resolveSeverity(c.Key, rules)
		msg := formatMessage(c)
		alerts = append(alerts, Alert{
			Key:      c.Key,
			Severity: sev,
			Message:  msg,
		})
	}
	return alerts
}

// Render formats a list of alerts into a human-readable string.
func Render(alerts []Alert) string {
	if len(alerts) == 0 {
		return "No alerts.\n"
	}
	var sb strings.Builder
	sb.WriteString("=== Drift Alerts ===\n")
	for _, a := range alerts {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", a.Severity, a.Key, a.Message))
	}
	return sb.String()
}

func resolveSeverity(key string, rules Rules) Severity {
	lower := strings.ToLower(key)
	for prefix, sev := range rules {
		if strings.Contains(lower, strings.ToLower(prefix)) {
			return sev
		}
	}
	return SeverityInfo
}

func formatMessage(c diff.Change) string {
	switch c.Type {
	case diff.Added:
		return fmt.Sprintf("key added with value %q", c.NewValue)
	case diff.Removed:
		return fmt.Sprintf("key removed (was %q)", c.OldValue)
	case diff.Modified:
		return fmt.Sprintf("value changed from %q to %q", c.OldValue, c.NewValue)
	default:
		return "unknown change"
	}
}
