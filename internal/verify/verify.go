package verify

import (
	"fmt"
	"strings"

	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/snapshot"
)

// Result holds the outcome of a verification check.
type Result struct {
	File     string
	Passed   bool
	Messages []string
}

// Options controls which checks are performed.
type Options struct {
	// RequireKeys fails verification if any of these keys are missing.
	RequireKeys []string
	// ForbidKeys fails verification if any of these keys are present.
	ForbidKeys []string
	// MaxChanges fails verification if the number of changes exceeds this value.
	// Zero means no limit.
	MaxChanges int
}

// Check compares a current snapshot against a previous one and applies
// the provided options to produce a Result.
func Check(prev, curr snapshot.Snapshot, opts Options) Result {
	changes := diff.Compare(prev, curr)

	result := Result{
		File:   curr.File,
		Passed: true,
	}

	if opts.MaxChanges > 0 && len(changes) > opts.MaxChanges {
		result.Passed = false
		result.Messages = append(result.Messages,
			fmt.Sprintf("change count %d exceeds maximum %d", len(changes), opts.MaxChanges))
	}

	keyIndex := make(map[string]string, len(curr.Values))
	for k, v := range curr.Values {
		keyIndex[k] = v
	}

	for _, req := range opts.RequireKeys {
		if _, ok := keyIndex[req]; !ok {
			result.Passed = false
			result.Messages = append(result.Messages,
				fmt.Sprintf("required key %q is missing", req))
		}
	}

	for _, forbid := range opts.ForbidKeys {
		if _, ok := keyIndex[forbid]; ok {
			result.Passed = false
			result.Messages = append(result.Messages,
				fmt.Sprintf("forbidden key %q is present", forbid))
		}
	}

	if result.Passed {
		result.Messages = []string{"all checks passed"}
	}

	return result
}

// Render formats a Result as a human-readable string.
func Render(r Result) string {
	status := "PASS"
	if !r.Passed {
		status = "FAIL"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] %s\n", status, r.File)
	for _, msg := range r.Messages {
		fmt.Fprintf(&sb, "  - %s\n", msg)
	}
	return sb.String()
}
