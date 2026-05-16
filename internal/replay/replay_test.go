package replay

import (
	"fmt"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/snapshot"
)

func makeLog(envs []string, times []time.Time, paths []string) history.Log {
	var entries []history.Entry
	for i, env := range envs {
		entries = append(entries, history.Entry{
			Environment:  env,
			Timestamp:    times[i],
			SnapshotPath: paths[i],
		})
	}
	return history.Log{Entries: entries}
}

func stubLoader(data map[string]snapshot.Snapshot) func(string) (snapshot.Snapshot, error) {
	return func(path string) (snapshot.Snapshot, error) {
		if s, ok := data[path]; ok {
			return s, nil
		}
		return snapshot.Snapshot{}, fmt.Errorf("not found: %s", path)
	}
}

func TestBuild_ReturnsFramesInOrder(t *testing.T) {
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)

	snaps := map[string]snapshot.Snapshot{
		"snap1.json": {Data: map[string]string{"a": "1"}},
		"snap2.json": {Data: map[string]string{"a": "2"}},
	}
	log := makeLog(
		[]string{"prod", "prod"},
		[]time.Time{t2, t1},
		[]string{"snap2.json", "snap1.json"},
	)

	r, err := Build(log, "prod", stubLoader(snaps))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(r.Frames))
	}
	if r.Frames[0].Snapshot.Data["a"] != "1" {
		t.Errorf("expected first frame to have a=1, got %s", r.Frames[0].Snapshot.Data["a"])
	}
}

func TestBuild_EmptyWhenNoMatchingEnv(t *testing.T) {
	log := makeLog([]string{"staging"}, []time.Time{time.Now()}, []string{"s.json"})
	r, err := Build(log, "prod", stubLoader(map[string]snapshot.Snapshot{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Frames) != 0 {
		t.Errorf("expected 0 frames, got %d", len(r.Frames))
	}
}

func TestBuild_ErrorOnMissingSnapshot(t *testing.T) {
	log := makeLog([]string{"prod"}, []time.Time{time.Now()}, []string{"missing.json"})
	_, err := Build(log, "prod", stubLoader(map[string]snapshot.Snapshot{}))
	if err == nil {
		t.Error("expected error for missing snapshot, got nil")
	}
}

func TestBuild_ErrorOnEmptyEnvironment(t *testing.T) {
	_, err := Build(history.Log{}, "", stubLoader(nil))
	if err == nil {
		t.Error("expected error for empty environment")
	}
}

func TestRender_ContainsEnvironmentAndFrames(t *testing.T) {
	r := Result{
		Environment: "prod",
		Frames: []Frame{
			{Index: 1, Environment: "prod", Timestamp: time.Now(), Snapshot: snapshot.Snapshot{Data: map[string]string{"key": "val"}}},
		},
	}
	out := Render(r)
	if !contains(out, "prod") {
		t.Error("expected output to contain environment name")
	}
	if !contains(out, "key = val") {
		t.Error("expected output to contain key=val")
	}
}

func TestRender_EmptyResult(t *testing.T) {
	r := Result{Environment: "dev", Frames: []Frame{}}
	out := Render(r)
	if !contains(out, "no history") {
		t.Errorf("expected no-history message, got: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
