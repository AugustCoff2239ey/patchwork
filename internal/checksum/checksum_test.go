package checksum_test

import (
	"strings"
	"testing"

	"github.com/patchwork/internal/checksum"
	"github.com/patchwork/internal/snapshot"
)

func makeSnap(env string, data map[string]string) snapshot.Snapshot {
	return snapshot.Snapshot{
		Environment: env,
		Timestamp:   "2024-01-01T00:00:00Z",
		Data:        data,
	}
}

func TestCompute_ReturnsDeterministicChecksum(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"key1": "val1", "key2": "val2"})
	r1, err := checksum.Compute(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := checksum.Compute(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.Checksum != r2.Checksum {
		t.Errorf("expected identical checksums, got %q and %q", r1.Checksum, r2.Checksum)
	}
}

func TestCompute_OrderIndependent(t *testing.T) {
	a := makeSnap("staging", map[string]string{"alpha": "1", "beta": "2", "gamma": "3"})
	b := makeSnap("staging", map[string]string{"gamma": "3", "alpha": "1", "beta": "2"})
	ra, _ := checksum.Compute(a)
	rb, _ := checksum.Compute(b)
	if ra.Checksum != rb.Checksum {
		t.Errorf("expected order-independent checksums, got %q and %q", ra.Checksum, rb.Checksum)
	}
}

func TestCompute_DifferentDataProducesDifferentChecksum(t *testing.T) {
	a := makeSnap("prod", map[string]string{"key": "old"})
	b := makeSnap("prod", map[string]string{"key": "new"})
	ra, _ := checksum.Compute(a)
	rb, _ := checksum.Compute(b)
	if ra.Checksum == rb.Checksum {
		t.Error("expected different checksums for different data")
	}
}

func TestCompute_EmptyEnvironment_ReturnsError(t *testing.T) {
	snap := makeSnap("", map[string]string{"k": "v"})
	_, err := checksum.Compute(snap)
	if err == nil {
		t.Error("expected error for empty environment")
	}
}

func TestCompute_EmptyData_ReturnsError(t *testing.T) {
	snap := makeSnap("dev", map[string]string{})
	_, err := checksum.Compute(snap)
	if err == nil {
		t.Error("expected error for empty snapshot data")
	}
}

func TestEqual_SameChecksum_ReturnsTrue(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"x": "y"})
	ra, _ := checksum.Compute(snap)
	rb, _ := checksum.Compute(snap)
	if !checksum.Equal(ra, rb) {
		t.Error("expected Equal to return true for identical snapshots")
	}
}

func TestRender_ContainsExpectedFields(t *testing.T) {
	snap := makeSnap("prod", map[string]string{"db": "postgres"})
	r, _ := checksum.Compute(snap)
	out := checksum.Render(r)
	for _, want := range []string{"Environment", "Checksum", "Keys", "Captured"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected Render output to contain %q", want)
		}
	}
}
