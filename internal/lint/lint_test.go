package lint

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		CapturedAt:  time.Now(),
		Data:        data,
	}
}

func TestRun_AllPasses(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"host": "localhost", "port": "8080"})
	report := Run(snap, DefaultRules())
	if report.FailCount != 0 {
		t.Errorf("expected 0 failures, got %d", report.FailCount)
	}
	if report.PassCount != len(DefaultRules()) {
		t.Errorf("expected %d passes, got %d", len(DefaultRules()), report.PassCount)
	}
}

func TestRun_DetectsEmptyValues(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"host": "localhost", "port": ""})
	report := Run(snap, DefaultRules())
	found := false
	for _, r := range report.Results {
		if r.Rule == "no-empty-values" && !r.Passed {
			found = true
		}
	}
	if !found {
		t.Error("expected no-empty-values rule to fail")
	}
}

func TestRun_DetectsEmptySnapshot(t *testing.T) {
	snap := makeSnap("dev", map[string]string{})
	report := Run(snap, DefaultRules())
	found := false
	for _, r := range report.Results {
		if r.Rule == "has-entries" && !r.Passed {
			found = true
		}
	}
	if !found {
		t.Error("expected has-entries rule to fail for empty snapshot")
	}
}

func TestRun_SetsEnvironment(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"k": "v"})
	report := Run(snap, DefaultRules())
	if report.Environment != "prod" {
		t.Errorf("expected environment 'prod', got '%s'", report.Environment)
	}
}

func TestRender_ContainsPassAndFail(t *testing.T) {
	snap := makeSnap("dev", map[string]string{"key": ""})
	report := Run(snap, DefaultRules())
	output := Render(report)
	if !strings.Contains(output, "FAIL") {
		t.Error("expected rendered output to contain FAIL")
	}
	if !strings.Contains(output, "PASS") {
		t.Error("expected rendered output to contain PASS")
	}
}

func TestRender_ContainsEnvironment(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"k": "v"})
	report := Run(snap, DefaultRules())
	output := Render(report)
	if !strings.Contains(output, "staging") {
		t.Error("expected rendered output to contain environment name")
	}
}

func TestRun_CustomRule(t *testing.T) {
	customRule := Rule{
		Name:    "must-have-region",
		Message: "missing required key: region",
		Check: func(snap snapshot.Snapshot) bool {
			_, ok := snap.Data["region"]
			return ok
		},
	}
	snap := makeSnap("prod", map[string]string{"host": "localhost"})
	report := Run(snap, []Rule{customRule})
	if report.FailCount != 1 {
		t.Errorf("expected 1 failure, got %d", report.FailCount)
	}
}
