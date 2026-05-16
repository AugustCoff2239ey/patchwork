package signature_test

import (
	"strings"
	"testing"

	"github.com/patchwork/internal/signature"
	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		Timestamp:   "2024-06-01T00:00:00Z",
		Data:        data,
	}
}

func TestCompute_ReturnsDeterministicHash(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"a": "1", "b": "2"})
	r1, err := signature.Compute(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := signature.Compute(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.Hash != r2.Hash {
		t.Errorf("expected identical hashes, got %s and %s", r1.Hash, r2.Hash)
	}
}

func TestCompute_DifferentDataProducesDifferentHash(t *testing.T) {
	s1 := makeSnap("prod", map[string]string{"a": "1"})
	s2 := makeSnap("prod", map[string]string{"a": "2"})
	r1, _ := signature.Compute(s1)
	r2, _ := signature.Compute(s2)
	if r1.Hash == r2.Hash {
		t.Error("expected different hashes for different data")
	}
}

func TestCompute_OrderIndependent(t *testing.T) {
	s1 := makeSnap("staging", map[string]string{"x": "10", "y": "20"})
	s2 := makeSnap("staging", map[string]string{"y": "20", "x": "10"})
	r1, _ := signature.Compute(s1)
	r2, _ := signature.Compute(s2)
	if r1.Hash != r2.Hash {
		t.Error("expected same hash regardless of map iteration order")
	}
}

func TestCompute_EmptyEnvironment_ReturnsError(t *testing.T) {
	snap := makeSnap("", map[string]string{"k": "v"})
	_, err := signature.Compute(snap)
	if err == nil {
		t.Error("expected error for empty environment")
	}
}

func TestEqual_SameSnapshot(t *testing.T) {
	snap := makeSnap("dev", map[string]string{"port": "8080"})
	r1, _ := signature.Compute(snap)
	r2, _ := signature.Compute(snap)
	if !signature.Equal(r1, r2) {
		t.Error("expected Equal to return true for identical snapshots")
	}
}

func TestRender_ContainsExpectedFields(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"db": "postgres"})
	r, _ := signature.Compute(snap)
	out := signature.Render(r)
	for _, want := range []string{"prod", "sha256", "keys"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected Render output to contain %q", want)
		}
	}
}
