package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/patchwork/internal/snapshot"
)

// Result holds the computed signature for a snapshot.
type Result struct {
	Environment string `json:"environment"`
	Timestamp   string `json:"timestamp"`
	Hash        string `json:"hash"`
	KeyCount    int    `json:"key_count"`
}

// Compute generates a deterministic SHA-256 hash over the snapshot's
// sorted key-value pairs, making it suitable for change detection.
func Compute(snap snapshot.Snapshot) (Result, error) {
	if snap.Environment == "" {
		return Result{}, fmt.Errorf("signature: snapshot environment must not be empty")
	}

	keys := make([]string, 0, len(snap.Data))
	for k := range snap.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+snap.Data[k])
	}

	raw, err := json.Marshal(pairs)
	if err != nil {
		return Result{}, fmt.Errorf("signature: failed to marshal pairs: %w", err)
	}

	sum := sha256.Sum256(raw)
	return Result{
		Environment: snap.Environment,
		Timestamp:   snap.Timestamp,
		Hash:        hex.EncodeToString(sum[:]),
		KeyCount:    len(snap.Data),
	}, nil
}

// Equal returns true when two Results share the same hash.
func Equal(a, b Result) bool {
	return a.Hash == b.Hash
}

// Render formats a Result for human-readable output.
func Render(r Result) string {
	return fmt.Sprintf("environment : %s\ntimestamp   : %s\nkeys        : %d\nsha256      : %s\n",
		r.Environment, r.Timestamp, r.KeyCount, r.Hash)
}
