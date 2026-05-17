package snapshot

import (
	"testing"
	"time"
)

func baseChainSnap(env string, ts time.Time, data map[string]string) Snapshot {
	return Snapshot{
		Environment: env,
		FilePath:    "/etc/cfg.json",
		CapturedAt:  ts,
		Data:        data,
	}
}

func TestBuildChain_OrdersChronologically(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		baseChainSnap("prod", now.Add(2*time.Hour), map[string]string{"k": "c"}),
		baseChainSnap("prod", now, map[string]string{"k": "a"}),
		baseChainSnap("prod", now.Add(time.Hour), map[string]string{"k": "b"}),
	}
	chain, err := BuildChain("prod", snaps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", chain.Len())
	}
	for i := 1; i < chain.Len(); i++ {
		if !chain.Entries[i-1].Timestamp.Before(chain.Entries[i].Timestamp) {
			t.Errorf("entries not in chronological order at index %d", i)
		}
	}
}

func TestBuildChain_EnvironmentMismatch(t *testing.T) {
	snaps := []Snapshot{
		baseChainSnap("prod", time.Now(), map[string]string{"k": "v"}),
		baseChainSnap("staging", time.Now(), map[string]string{"k": "v"}),
	}
	_, err := BuildChain("prod", snaps)
	if err == nil {
		t.Fatal("expected error for environment mismatch, got nil")
	}
}

func TestBuildChain_EmptyEnvironment(t *testing.T) {
	_, err := BuildChain("", []Snapshot{})
	if err == nil {
		t.Fatal("expected error for empty environment")
	}
}

func TestChain_Latest_ReturnsNewest(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		baseChainSnap("dev", now, map[string]string{"x": "1"}),
		baseChainSnap("dev", now.Add(time.Hour), map[string]string{"x": "2"}),
	}
	chain, _ := BuildChain("dev", snaps)
	latest, err := chain.Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !latest.Timestamp.Equal(now.Add(time.Hour)) {
		t.Errorf("expected latest timestamp %v, got %v", now.Add(time.Hour), latest.Timestamp)
	}
}

func TestChain_Latest_EmptyChain(t *testing.T) {
	chain := Chain{Environment: "prod", Entries: nil}
	_, err := chain.Latest()
	if err == nil {
		t.Fatal("expected error for empty chain")
	}
}

func TestRenderChain_ContainsEnvironment(t *testing.T) {
	chain, _ := BuildChain("staging", []Snapshot{
		baseChainSnap("staging", time.Now(), map[string]string{"a": "1"}),
	})
	out := RenderChain(chain)
	if len(out) == 0 {
		t.Fatal("expected non-empty render output")
	}
	if !containsStr(out, "staging") {
		t.Errorf("expected render to contain environment name")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
