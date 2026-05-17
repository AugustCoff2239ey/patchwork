package scope

import (
	"strings"
	"testing"
	"time"

	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		Timestamp:   time.Now(),
		Data:        data,
	}
}

func TestApply_MatchesAllKeys(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"host": "localhost", "port": "8080"})
	sc := Scope{Name: "network", Environment: "prod", Keys: []string{"host", "port"}}

	r, err := Apply(snap, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Matched) != 2 {
		t.Errorf("expected 2 matched keys, got %d", len(r.Matched))
	}
	if len(r.Missing) != 0 {
		t.Errorf("expected no missing keys, got %v", r.Missing)
	}
}

func TestApply_DetectsMissingKeys(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"host": "localhost"})
	sc := Scope{Name: "network", Environment: "prod", Keys: []string{"host", "port", "timeout"}}

	r, err := Apply(snap, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Matched) != 1 {
		t.Errorf("expected 1 matched key, got %d", len(r.Matched))
	}
	if len(r.Missing) != 2 {
		t.Errorf("expected 2 missing keys, got %v", r.Missing)
	}
}

func TestApply_EnvironmentMismatch_ReturnsError(t *testing.T) {
	snap := makeSnap("staging", map[string]string{"host": "localhost"})
	sc := Scope{Name: "network", Environment: "prod", Keys: []string{"host"}}

	_, err := Apply(snap, sc)
	if err == nil {
		t.Fatal("expected error for environment mismatch")
	}
}

func TestApply_EmptyScopeName_ReturnsError(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"host": "localhost"})
	sc := Scope{Name: "", Environment: "prod", Keys: []string{"host"}}

	_, err := Apply(snap, sc)
	if err == nil {
		t.Fatal("expected error for empty scope name")
	}
}

func TestRender_ContainsExpectedSections(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"host": "localhost", "port": "8080"})
	sc := Scope{Name: "network", Environment: "prod", Keys: []string{"host", "port", "missing_key"}}

	r, _ := Apply(snap, sc)
	out := Render(r)

	for _, want := range []string{"network", "prod", "host", "localhost", "missing_key"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}
