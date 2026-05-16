package validate_test

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
	"github.com/patchwork/internal/validate"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{Environment: env, CapturedAt: time.Now(), Data: data}
}

func TestRun_AllPasses(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"db_host": "localhost", "port": "5432"})
	rules := []validate.Rule{
		{Key: "db_host", Required: true},
		{Key: "port", Pattern: `^\d+$`, Required: true},
	}
	report := validate.Run(snap, rules)
	if report.Failed != 0 {
		t.Errorf("expected 0 failures, got %d", report.Failed)
	}
	if report.Passed != 2 {
		t.Errorf("expected 2 passes, got %d", report.Passed)
	}
}

func TestRun_MissingRequiredKey(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"port": "8080"})
	rules := []validate.Rule{
		{Key: "db_host", Required: true},
	}
	report := validate.Run(snap, rules)
	if report.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", report.Failed)
	}
	if !strings.Contains(report.Results[0].Message, "missing") {
		t.Errorf("expected missing message, got: %s", report.Results[0].Message)
	}
}

func TestRun_PatternMismatch(t *testing.T) {
	snap := makeSnap("dev", map[string]string{"port": "not-a-number"})
	rules := []validate.Rule{
		{Key: "port", Pattern: `^\d+$`, Required: true},
	}
	report := validate.Run(snap, rules)
	if report.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", report.Failed)
	}
}

func TestRun_OptionalMissingKey_Passes(t *testing.T) {
	snap := makeSnap("dev", map[string]string{})
	rules := []validate.Rule{
		{Key: "optional_flag", Required: false},
	}
	report := validate.Run(snap, rules)
	if report.Failed != 0 {
		t.Errorf("expected 0 failures for optional missing key, got %d", report.Failed)
	}
}

func TestRender_ContainsEnvironmentAndStatus(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"key": "val"})
	rules := []validate.Rule{{Key: "key", Required: true}}
	report := validate.Run(snap, rules)
	out := validate.Render(report)
	if !strings.Contains(out, "prod") {
		t.Error("expected environment name in render output")
	}
	if !strings.Contains(out, "PASS") {
		t.Error("expected PASS in render output")
	}
}
