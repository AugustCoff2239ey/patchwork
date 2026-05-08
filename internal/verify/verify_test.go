package verify_test

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
	"github.com/patchwork/internal/verify"
)

func makeSnap(file string, values map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		File:      file,
		CapturedAt: time.Now(),
		Values:    values,
	}
}

func TestCheck_PassesWithNoChangesAndNoOptions(t *testing.T) {
	prev := makeSnap("app.conf", map[string]string{"a": "1"})
	curr := makeSnap("app.conf", map[string]string{"a": "1"})

	r := verify.Check(prev, curr, verify.Options{})

	if !r.Passed {
		t.Errorf("expected pass, got messages: %v", r.Messages)
	}
}

func TestCheck_FailsWhenMaxChangesExceeded(t *testing.T) {
	prev := makeSnap("app.conf", map[string]string{"a": "1", "b": "2"})
	curr := makeSnap("app.conf", map[string]string{"a": "99", "b": "99"})

	r := verify.Check(prev, curr, verify.Options{MaxChanges: 1})

	if r.Passed {
		t.Error("expected failure due to too many changes")
	}
	if !containsMsg(r.Messages, "exceeds maximum") {
		t.Errorf("expected 'exceeds maximum' message, got: %v", r.Messages)
	}
}

func TestCheck_FailsWhenRequiredKeyMissing(t *testing.T) {
	prev := makeSnap("app.conf", map[string]string{})
	curr := makeSnap("app.conf", map[string]string{"host": "localhost"})

	r := verify.Check(prev, curr, verify.Options{RequireKeys: []string{"port"}})

	if r.Passed {
		t.Error("expected failure for missing required key")
	}
	if !containsMsg(r.Messages, "required key") {
		t.Errorf("expected 'required key' message, got: %v", r.Messages)
	}
}

func TestCheck_FailsWhenForbiddenKeyPresent(t *testing.T) {
	prev := makeSnap("app.conf", map[string]string{})
	curr := makeSnap("app.conf", map[string]string{"debug": "true"})

	r := verify.Check(prev, curr, verify.Options{ForbidKeys: []string{"debug"}})

	if r.Passed {
		t.Error("expected failure for forbidden key")
	}
	if !containsMsg(r.Messages, "forbidden key") {
		t.Errorf("expected 'forbidden key' message, got: %v", r.Messages)
	}
}

func TestCheck_PassesWhenRequiredKeyPresent(t *testing.T) {
	prev := makeSnap("app.conf", map[string]string{})
	curr := makeSnap("app.conf", map[string]string{"port": "8080"})

	r := verify.Check(prev, curr, verify.Options{RequireKeys: []string{"port"}})

	if !r.Passed {
		t.Errorf("expected pass, got: %v", r.Messages)
	}
}

func TestRender_ContainsStatusAndFile(t *testing.T) {
	r := verify.Result{File: "app.conf", Passed: false, Messages: []string{"something wrong"}}
	out := verify.Render(r)

	if !strings.Contains(out, "FAIL") {
		t.Error("expected FAIL in output")
	}
	if !strings.Contains(out, "app.conf") {
		t.Error("expected filename in output")
	}
	if !strings.Contains(out, "something wrong") {
		t.Error("expected message in output")
	}
}

func containsMsg(msgs []string, substr string) bool {
	for _, m := range msgs {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}
