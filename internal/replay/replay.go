package replay

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

// Frame represents a single step in a replay sequence.
type Frame struct {
	Index       int
	Environment string
	Timestamp   time.Time
	Snapshot    snapshot.Snapshot
}

// Result holds the ordered frames produced by a replay.
type Result struct {
	Environment string
	Frames      []Frame
}

// Build constructs a replay sequence for the given environment from history entries.
// It loads each snapshot referenced by the log and returns them in chronological order.
func Build(log history.Log, env string, loadFn func(path string) (snapshot.Snapshot, error)) (Result, error) {
	if env == "" {
		return Result{}, fmt.Errorf("replay: environment must not be empty")
	}

	var entries []history.Entry
	for _, e := range log.Entries {
		if e.Environment == env {
			entries = append(entries, e)
		}
	}

	if len(entries) == 0 {
		return Result{Environment: env, Frames: []Frame{}}, nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	var frames []Frame
	for idx, e := range entries {
		snap, err := loadFn(e.SnapshotPath)
		if err != nil {
			return Result{}, fmt.Errorf("replay: loading snapshot %q: %w", e.SnapshotPath, err)
		}
		frames = append(frames, Frame{
			Index:       idx + 1,
			Environment: e.Environment,
			Timestamp:   e.Timestamp,
			Snapshot:    snap,
		})
	}

	return Result{Environment: env, Frames: frames}, nil
}

// Render formats a replay Result as a human-readable string.
func Render(r Result) string {
	if len(r.Frames) == 0 {
		return fmt.Sprintf("replay: no history found for environment %q\n", r.Environment)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Replay for environment: %s (%d frames)\n", r.Environment, len(r.Frames))
	fmt.Fprintln(&sb, strings.Repeat("-", 48))

	for _, f := range r.Frames {
		fmt.Fprintf(&sb, "[%d] %s\n", f.Index, f.Timestamp.Format(time.RFC3339))
		keys := make([]string, 0, len(f.Snapshot.Data))
		for k := range f.Snapshot.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&sb, "    %s = %s\n", k, f.Snapshot.Data[k])
		}
	}

	return sb.String()
}
