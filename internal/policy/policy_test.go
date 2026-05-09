package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/patchwork/internal/diff"
)

func writePolicy(t *testing.T, pf PolicyFile) string {
	t.Helper()
	data, err := json.Marshal(pf)
	if err != nil {
		t.Fatalf("marshal policy: %v", err)
	}
	path := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	return path
}

func makeChanges(keys ...string) []diff.Change {
	var changes []diff.Change
	for _, k := range keys {
		changes = append(changes, diff.Change{Key: k, Type: diff.Modified, Old: "a", New: "b"})
	}
	return changes
}

func TestEvaluate_MaxChangesViolation(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "limit", MaxChanges: 2}}}
	changes := makeChanges("a", "b", "c")
	violations := Evaluate(pf, "prod", changes)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Rule != "limit" {
		t.Errorf("expected rule 'limit', got %q", violations[0].Rule)
	}
}

func TestEvaluate_ForbidKeyViolation(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "no-secret", ForbidKeys: []string{"SECRET"}}}}
	changes := makeChanges("SECRET", "HOST")
	violations := Evaluate(pf, "prod", changes)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestEvaluate_RequireKeyViolation(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "must-version", RequireKeys: []string{"VERSION"}}}}
	changes := makeChanges("HOST")
	violations := Evaluate(pf, "prod", changes)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestEvaluate_EnvFilter_SkipsNonMatchingEnv(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "prod-only", Environments: []string{"prod"}, MaxChanges: 1}}}
	changes := makeChanges("a", "b", "c")
	violations := Evaluate(pf, "staging", changes)
	if len(violations) != 0 {
		t.Errorf("expected no violations for staging, got %d", len(violations))
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "limit", MaxChanges: 10}}}
	changes := makeChanges("a", "b")
	violations := Evaluate(pf, "prod", changes)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestLoad_ParsesFile(t *testing.T) {
	pf := PolicyFile{Rules: []Rule{{Name: "r1", MaxChanges: 5}}}
	path := writePolicy(t, pf)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Rules) != 1 || loaded.Rules[0].Name != "r1" {
		t.Errorf("unexpected loaded policy: %+v", loaded)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/policy.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRender_NoViolations(t *testing.T) {
	out := Render(nil)
	if out != "policy: all rules passed\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRender_WithViolations(t *testing.T) {
	v := []Violation{{Rule: "r1", Message: "too many changes"}}
	out := Render(v)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}
