package drift

import (
	"fmt"
	"sort"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/snapshot"
)

// Trend represents the drift trend for a single environment.
type Trend struct {
	Environment string
	Period      string
	TotalDiffs  int
	AvgPerDay   float64
	Peak        time.Time
	PeakCount   int
}

// Report holds drift trends across all environments.
type Report struct {
	GeneratedAt time.Time
	Trends      []Trend
}

// Analyze computes drift trends from history log entries over the given number of days.
func Analyze(log history.Log, days int) Report {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	envDays := map[string]map[string]int{}

	for _, entry := range log.Entries {
		if entry.Timestamp.Before(cutoff) {
			continue
		}
		env := entry.Environment
		day := entry.Timestamp.Format("2006-01-02")
		if envDays[env] == nil {
			envDays[env] = map[string]int{}
		}

		snap, err := snapshot.Load(entry.SnapshotPath)
		if err != nil {
			continue
		}
		changes := diff.Compare(snapshot.Snapshot{}, snap)
		envDays[env][day] += len(changes)
	}

	var trends []Trend
	for env, dayMap := range envDays {
		total := 0
		peak := ""
		peakCount := 0
		for d, c := range dayMap {
			total += c
			if c > peakCount {
				peakCount = c
				peak = d
			}
		}
		peakTime, _ := time.Parse("2006-01-02", peak)
		avg := 0.0
		if days > 0 {
			avg = float64(total) / float64(days)
		}
		trends = append(trends, Trend{
			Environment: env,
			Period:      fmt.Sprintf("%d days", days),
			TotalDiffs:  total,
			AvgPerDay:   avg,
			Peak:        peakTime,
			PeakCount:   peakCount,
		})
	}
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Environment < trends[j].Environment
	})
	return Report{GeneratedAt: time.Now().UTC(), Trends: trends}
}

// Render formats a drift Report as a human-readable string.
func Render(r Report) string {
	if len(r.Trends) == 0 {
		return "No drift data available.\n"
	}
	out := fmt.Sprintf("Drift Trend Report — %s\n", r.GeneratedAt.Format("2006-01-02 15:04 UTC"))
	out += fmt.Sprintf("%-20s %-12s %10s %12s %12s\n", "Environment", "Period", "TotalDiffs", "Avg/Day", "PeakCount")
	out += fmt.Sprintf("%s\n", fmt.Sprintf("%s", "─────────────────────────────────────────────────────────────"))
	for _, t := range r.Trends {
		out += fmt.Sprintf("%-20s %-12s %10d %12.2f %12d\n",
			t.Environment, t.Period, t.TotalDiffs, t.AvgPerDay, t.PeakCount)
	}
	return out
}
