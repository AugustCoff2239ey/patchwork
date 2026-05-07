package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/user/patchwork/internal/snapshot"
)

// Baseline represents a pinned reference snapshot for an environment.
type Baseline struct {
	Environment string            `json:"environment"`
	PinnedAt    time.Time         `json:"pinned_at"`
	Snapshot    snapshot.Snapshot `json:"snapshot"`
}

// Pin creates a new Baseline from the given snapshot.
func Pin(env string, snap snapshot.Snapshot) Baseline {
	return Baseline{
		Environment: env,
		PinnedAt:    time.Now().UTC(),
		Snapshot:    snap,
	}
}

// Save writes the baseline to a JSON file at the given path.
func Save(path string, b Baseline) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// Load reads a baseline from a JSON file at the given path.
func Load(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Baseline{}, fmt.Errorf("baseline: no baseline found at %s", path)
		}
		return Baseline{}, fmt.Errorf("baseline: read %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return b, nil
}
