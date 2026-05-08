package redact

import (
	"regexp"
	"strings"
)

// DefaultPatterns are common sensitive key patterns to redact.
var DefaultPatterns = []string{
	"password",
	"secret",
	"token",
	"api_key",
	"apikey",
	"private_key",
	"credential",
}

const redactedValue = "[REDACTED]"

// Options configures redaction behavior.
type Options struct {
	// Patterns is a list of key substrings (case-insensitive) to redact.
	Patterns []string
}

// DefaultOptions returns Options populated with DefaultPatterns.
func DefaultOptions() Options {
	return Options{Patterns: DefaultPatterns}
}

// Apply returns a copy of the snapshot map with sensitive values replaced.
func Apply(data map[string]string, opts Options) map[string]string {
	patterns := opts.Patterns
	if len(patterns) == 0 {
		patterns = DefaultPatterns
	}

	// Build a single case-insensitive regexp from all patterns.
	parts := make([]string, len(patterns))
	for i, p := range patterns {
		parts[i] = regexp.QuoteMeta(strings.ToLower(p))
	}
	combined := regexp.MustCompile(strings.Join(parts, "|"))

	result := make(map[string]string, len(data))
	for k, v := range data {
		if combined.MatchString(strings.ToLower(k)) {
			result[k] = redactedValue
		} else {
			result[k] = v
		}
	}
	return result
}

// IsSensitive reports whether a key matches any of the provided patterns.
func IsSensitive(key string, patterns []string) bool {
	lower := strings.ToLower(key)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
