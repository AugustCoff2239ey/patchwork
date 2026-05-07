package alert_test

import (
	"strings"
	"testing"

	"github.com/patchwork/internal/alert"
	"github.com/patchwork/internal/diff"
)

func makeChanges() []diff.Change {
	return []diff.Change{
		{Key: "db_password", Type: diff.Modified, OldValue: "old", NewValue: "new"},
		{Key: "app_host", Type: diff.Modified, OldValue: "localhost", NewValue: "prod.example.com"},
		{Key: "log_level", Type: diff.Added, NewValue: "debug"},
		{Key: "deprecated_key", Type: diff.Removed, OldValue: "value"},
	}
}

func TestEvaluate_AssignsCriticalSeverity(t *testing.T) {
	changes := makeChanges()
	alerts := alert.Evaluate(changes, alert.DefaultRules)

	for _, a := range alerts {
		if a.Key == "db_password" && a.Severity != alert.SeverityCritical {
			t.Errorf("expected CRITICAL for db_password, got %s", a.Severity)
		}
	}
}

func TestEvaluate_AssignsWarningSeverity(t *testing.T) {
	changes := makeChanges()
	alerts := alert.Evaluate(changes, alert.DefaultRules)

	for _, a := range alerts {
		if a.Key == "app_host" && a.Severity != alert.SeverityWarning {
			t.Errorf("expected WARNING for app_host, got %s", a.Severity)
		}
	}
}

func TestEvaluate_DefaultsToInfo(t *testing.T) {
	changes := makeChanges()
	alerts := alert.Evaluate(changes, alert.DefaultRules)

	for _, a := range alerts {
		if a.Key == "log_level" && a.Severity != alert.SeverityInfo {
			t.Errorf("expected INFO for log_level, got %s", a.Severity)
		}
	}
}

func TestEvaluate_ReturnsOneAlertPerChange(t *testing.T) {
	changes := makeChanges()
	alerts := alert.Evaluate(changes, alert.DefaultRules)
	if len(alerts) != len(changes) {
		t.Errorf("expected %d alerts, got %d", len(changes), len(alerts))
	}
}

func TestRender_NoAlerts(t *testing.T) {
	out := alert.Render(nil)
	if !strings.Contains(out, "No alerts") {
		t.Errorf("expected 'No alerts' message, got: %s", out)
	}
}

func TestRender_ContainsSeverityAndKey(t *testing.T) {
	alerts := []alert.Alert{
		{Key: "api_token", Severity: alert.SeverityCritical, Message: "key added with value \"abc\""},
	}
	out := alert.Render(alerts)
	if !strings.Contains(out, "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got: %s", out)
	}
	if !strings.Contains(out, "api_token") {
		t.Errorf("expected api_token in output, got: %s", out)
	}
}
