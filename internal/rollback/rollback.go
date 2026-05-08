package rollback

import (
	"fmt"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

// Plan describes a rollback operation to be performed.
type Plan struct {
	Environment string
	FromEntry   history.Entry
	ToEntry     history.Entry
	CreatedAt   time.Time
}

// Result captures the outcome of a rollback.
type Result struct {
	Plan      Plan
	Applied   bool
	Message   string
}

// Prepare builds a rollback plan from the two most recent history entries
// for the given environment. Returns an error if fewer than two entries exist.
func Prepare(env string, log history.Log) (Plan, error) {
	entries := make([]history.Entry, 0)
	for _, e := range log.Entries {
		if e.Environment == env {
			entries = append(entries, e)
		}
	}
	if len(entries) < 2 {
		return Plan{}, fmt.Errorf("rollback: need at least 2 history entries for environment %q, found %d", env, len(entries))
	}
	// entries are assumed sorted ascending; take last two
	from := entries[len(entries)-1]
	to := entries[len(entries)-2]
	return Plan{
		Environment: env,
		FromEntry:   from,
		ToEntry:     to,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// Apply executes the rollback plan by loading the target snapshot and
// saving it as the new current snapshot at the destination path.
func Apply(plan Plan, destPath string) (Result, error) {
	snap, err := snapshot.Load(plan.ToEntry.SnapshotPath)
	if err != nil {
		return Result{Plan: plan}, fmt.Errorf("rollback: load snapshot: %w", err)
	}
	if err := snapshot.Save(snap, destPath); err != nil {
		return Result{Plan: plan}, fmt.Errorf("rollback: save snapshot: %w", err)
	}
	return Result{
		Plan:    plan,
		Applied: true,
		Message: fmt.Sprintf("rolled back %s from %s to %s", plan.Environment, plan.FromEntry.CapturedAt.Format(time.RFC3339), plan.ToEntry.CapturedAt.Format(time.RFC3339)),
	}, nil
}
