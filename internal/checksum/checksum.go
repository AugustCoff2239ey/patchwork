package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/patchwork/internal/snapshot"
)

// Result holds the checksum for a snapshot and metadata about its computation.
type Result struct {
	Environment string
	Checksum    string
	KeyCount    int
	Timestamp   string
}

// Compute generates a deterministic SHA-256 checksum for a snapshot's data.
// Keys are sorted before hashing to ensure order-independence.
func Compute(snap snapshot.Snapshot) (Result, error) {
	if snap.Environment == "" {
		return Result{}, fmt.Errorf("checksum: environment must not be empty")
	}
	if len(snap.Data) == 0 {
		return Result{}, fmt.Errorf("checksum: snapshot data is empty for environment %q", snap.Environment)
	}

	keys := make([]string, 0, len(snap.Data))
	for k := range snap.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%s\n", k, snap.Data[k])
	}

	sum := hex.EncodeToString(h.Sum(nil))
	return Result{
		Environment: snap.Environment,
		Checksum:    sum,
		KeyCount:    len(snap.Data),
		Timestamp:   snap.Timestamp,
	}, nil
}

// Equal returns true if two results share the same checksum value.
func Equal(a, b Result) bool {
	return a.Checksum == b.Checksum
}

// Render formats a Result as a human-readable string.
func Render(r Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Environment : %s\n", r.Environment))
	sb.WriteString(fmt.Sprintf("Checksum    : %s\n", r.Checksum))
	sb.WriteString(fmt.Sprintf("Keys        : %d\n", r.KeyCount))
	if r.Timestamp != "" {
		sb.WriteString(fmt.Sprintf("Captured    : %s\n", r.Timestamp))
	}
	return sb.String()
}
