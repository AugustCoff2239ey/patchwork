package redact_test

import (
	"testing"

	"github.com/user/patchwork/internal/redact"
)

func baseData() map[string]string {
	return map[string]string{
		"db_password":  "s3cr3t",
		"api_key":      "abc123",
		"service_token": "tok-xyz",
		"host":         "localhost",
		"port":         "5432",
		"DB_SECRET":    "topsecret",
	}
}

func TestApply_RedactsSensitiveKeys(t *testing.T) {
	result := redact.Apply(baseData(), redact.DefaultOptions())

	sensitive := []string{"db_password", "api_key", "service_token", "DB_SECRET"}
	for _, k := range sensitive {
		if result[k] != "[REDACTED]" {
			t.Errorf("expected key %q to be redacted, got %q", k, result[k])
		}
	}
}

func TestApply_PreservesNonSensitiveKeys(t *testing.T) {
	result := redact.Apply(baseData(), redact.DefaultOptions())

	if result["host"] != "localhost" {
		t.Errorf("expected host to be preserved, got %q", result["host"])
	}
	if result["port"] != "5432" {
		t.Errorf("expected port to be preserved, got %q", result["port"])
	}
}

func TestApply_CustomPatterns(t *testing.T) {
	data := map[string]string{
		"internal_endpoint": "http://internal",
		"public_url":        "http://public",
	}
	opts := redact.Options{Patterns: []string{"internal"}}
	result := redact.Apply(data, opts)

	if result["internal_endpoint"] != "[REDACTED]" {
		t.Errorf("expected internal_endpoint to be redacted")
	}
	if result["public_url"] != "http://public" {
		t.Errorf("expected public_url to be preserved")
	}
}

func TestApply_EmptyPatternsFallsBackToDefaults(t *testing.T) {
	data := map[string]string{"password": "hunter2", "name": "alice"}
	result := redact.Apply(data, redact.Options{Patterns: nil})

	if result["password"] != "[REDACTED]" {
		t.Errorf("expected password to be redacted with default patterns")
	}
	if result["name"] != "alice" {
		t.Errorf("expected name to be preserved")
	}
}

func TestIsSensitive_MatchesCaseInsensitive(t *testing.T) {
	if !redact.IsSensitive("DB_PASSWORD", redact.DefaultPatterns) {
		t.Error("expected DB_PASSWORD to be sensitive")
	}
	if redact.IsSensitive("hostname", redact.DefaultPatterns) {
		t.Error("expected hostname to not be sensitive")
	}
}
