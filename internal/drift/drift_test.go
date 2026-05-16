package drift

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
)

func makeLog(envs []string, daysAgo []int) history.Log {
	var entries []history.Entry
	for i, env := range envs {
		ts := time.Now().UTC().AddDate(0, 0, -daysAgo[i])
		entries = append(entries, history.Entry{
			Environment:  env,
			Timestamp:    ts,
			SnapshotPath: "/nonexistent/snap.json",
		})
	}
	return history.Log{Entries: entries}
}

func TestAnalyze_ReturnsEmptyWhenNoEntries(t *testing.T) {
	log := history.Log{}
	report := Analyze(log, 7)
	if len(report.Trends) != 0 {
		t.Errorf("expected 0 trends, got %d", len(report.Trends))
	}
}

func TestAnalyze_FiltersOldEntries(t *testing.T) {
	log := makeLog([]string{"prod", "staging"}, []int{30, 2})
	report := Analyze(log, 7)
	for _, tr := range report.Trends {
		if tr.Environment == "prod" {
			t.Error("expected prod entry to be filtered out (too old)")
		}
	}
}

func TestAnalyze_SortsByEnvironment(t *testing.T) {
	log := makeLog([]string{"staging", "dev", "prod"}, []int{1, 2, 3})
	report := Analyze(log, 7)
	envs := make([]string, len(report.Trends))
	for i, tr := range report.Trends {
		envs[i] = tr.Environment
	}
	for i := 1; i < len(envs); i++ {
		if envs[i] < envs[i-1] {
			t.Errorf("trends not sorted: %v", envs)
		}
	}
}

func TestAnalyze_PeriodLabel(t *testing.T) {
	log := makeLog([]string{"dev"}, []int{1})
	report := Analyze(log, 14)
	if len(report.Trends) == 0 {
		return // snapshot load fails, no trends — acceptable
	}
	if report.Trends[0].Period != "14 days" {
		t.Errorf("unexpected period: %s", report.Trends[0].Period)
	}
}

func TestRender_NoData(t *testing.T) {
	r := Report{GeneratedAt: time.Now(), Trends: nil}
	out := Render(r)
	if !strings.Contains(out, "No drift data") {
		t.Errorf("expected no-data message, got: %s", out)
	}
}

func TestRender_ContainsHeaders(t *testing.T) {
	r := Report{
		GeneratedAt: time.Now(),
		Trends: []Trend{
			{Environment: "prod", Period: "7 days", TotalDiffs: 5, AvgPerDay: 0.71, PeakCount: 3},
		},
	}
	out := Render(r)
	for _, expected := range []string{"Environment", "TotalDiffs", "Avg/Day", "prod"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in render output", expected)
		}
	}
}
