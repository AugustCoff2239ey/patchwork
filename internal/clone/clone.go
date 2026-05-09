package clone

import (
	"fmt"
	"time"

	"github.com/patchwork/internal/snapshot"
)

// Result holds the outcome of a clone operation.
type Result struct {
	SourceEnvironment string
	TargetEnvironment string
	Snapshot          snapshot.Snapshot
	ClonedAt          time.Time
}

// Options controls how a clone is performed.
type Options struct {
	ExcludeKeys []string
	Overwrite   bool
}

// Apply clones a snapshot from one environment to another, optionally
// excluding specific keys. Returns an error if the target already exists
// and Overwrite is false.
func Apply(src snapshot.Snapshot, targetEnv string, opts Options) (Result, error) {
	if targetEnv == "" {
		return Result{}, fmt.Errorf("target environment must not be empty")
	}
	if src.Environment == targetEnv && !opts.Overwrite {
		return Result{}, fmt.Errorf("source and target environment are the same: %q", targetEnv)
	}

	excluded := make(map[string]bool, len(opts.ExcludeKeys))
	for _, k := range opts.ExcludeKeys {
		excluded[k] = true
	}

	cloned := make(map[string]string, len(src.Data))
	for k, v := range src.Data {
		if !excluded[k] {
			cloned[k] = v
		}
	}

	dst := snapshot.Snapshot{
		Environment: targetEnv,
		CapturedAt:  time.Now().UTC(),
		Data:        cloned,
	}

	return Result{
		SourceEnvironment: src.Environment,
		TargetEnvironment: targetEnv,
		Snapshot:          dst,
		ClonedAt:          dst.CapturedAt,
	}, nil
}

// Render returns a human-readable summary of the clone result.
func Render(r Result) string {
	return fmt.Sprintf(
		"Cloned %d key(s) from environment %q to %q at %s\n",
		len(r.Snapshot.Data),
		r.SourceEnvironment,
		r.TargetEnvironment,
		r.ClonedAt.Format(time.RFC3339),
	)
}
