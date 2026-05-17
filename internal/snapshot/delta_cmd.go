package snapshot

import (
	"fmt"
	"io"
	"strings"
)

// RenderDelta writes a human-readable summary of a DiffSnapshot to w.
func RenderDelta(w io.Writer, d DiffSnapshot) {
	fmt.Fprintf(w, "Delta for environment: %s\n", d.Environment)
	fmt.Fprintf(w, "From: %s\n", d.From.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "To:   %s\n", d.To.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w, strings.Repeat("-", 40))

	if d.IsEmpty() {
		fmt.Fprintln(w, "No changes detected.")
		return
	}

	if len(d.Added) > 0 {
		fmt.Fprintln(w, "Added:")
		for _, k := range sortedKeys(d.Added) {
			fmt.Fprintf(w, "  + %s = %s\n", k, d.Added[k])
		}
	}

	if len(d.Removed) > 0 {
		fmt.Fprintln(w, "Removed:")
		for _, k := range sortedKeys(d.Removed) {
			fmt.Fprintf(w, "  - %s = %s\n", k, d.Removed[k])
		}
	}

	if len(d.Modified) > 0 {
		fmt.Fprintln(w, "Modified:")
		for _, k := range sortedKeys(d.Modified) {
			fmt.Fprintf(w, "  ~ %s => %s\n", k, d.Modified[k])
		}
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// reuse sort from diff_snapshot.go via Keys-like logic
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
