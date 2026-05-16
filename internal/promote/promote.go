package promote

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Plan describes a promotion from one environment to another.
type Plan struct {
	SourceEnv string
	TargetEnv string
	Keys      []string
	Applied   int
	Skipped   int
}

// Options controls how promotion behaves.
type Options struct {
	OnlyKeys  []string // if non-empty, only promote these keys
	Overwrite bool     // overwrite keys that already exist in target
}

// Apply copies keys from src snapshot into a new snapshot based on dst,
// returning the resulting snapshot and a Plan describing what happened.
func Apply(src, dst snapshot.Snapshot, opts Options) (snapshot.Snapshot, Plan, error) {
	if src.Environment == "" {
		return snapshot.Snapshot{}, Plan{}, fmt.Errorf("source snapshot has no environment")
	}
	if dst.Environment == "" {
		return snapshot.Snapshot{}, Plan{}, fmt.Errorf("target snapshot has no environment")
	}
	if src.Environment == dst.Environment && !opts.Overwrite {
		return snapshot.Snapshot{}, Plan{}, fmt.Errorf("source and target environments are the same")
	}

	allowSet := map[string]bool{}
	for _, k := range opts.OnlyKeys {
		allowSet[k] = true
	}

	result := snapshot.Snapshot{
		Environment: dst.Environment,
		Timestamp:   dst.Timestamp,
		Data:        make(map[string]string),
	}
	for k, v := range dst.Data {
		result.Data[k] = v
	}

	plan := Plan{
		SourceEnv: src.Environment,
		TargetEnv: dst.Environment,
	}

	for k, v := range src.Data {
		if len(allowSet) > 0 && !allowSet[k] {
			continue
		}
		_, exists := result.Data[k]
		if exists && !opts.Overwrite {
			plan.Skipped++
			continue
		}
		result.Data[k] = v
		plan.Keys = append(plan.Keys, k)
		plan.Applied++
	}

	return result, plan, nil
}

// Render returns a human-readable summary of the promotion plan.
func Render(p Plan) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Promote: %s → %s\n", p.SourceEnv, p.TargetEnv)
	fmt.Fprintf(&sb, "  Applied : %d\n", p.Applied)
	fmt.Fprintf(&sb, "  Skipped : %d\n", p.Skipped)
	if len(p.Keys) > 0 {
		fmt.Fprintf(&sb, "  Keys    : %s\n", strings.Join(p.Keys, ", "))
	}
	return sb.String()
}
