package watch_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/watch"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.env")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestWatcher_NoDriftOnFirstPoll(t *testing.T) {
	path := writeTempConfig(t, "KEY=value\n")
	w := watch.New(watch.WatchConfig{
		FilePath:    path,
		Environment: "test",
		Interval:    time.Second,
	})

	event, err := w.Poll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != nil {
		t.Errorf("expected nil event on first poll, got %+v", event)
	}
}

func TestWatcher_DetectsDriftOnChange(t *testing.T) {
	path := writeTempConfig(t, "KEY=original\n")
	w := watch.New(watch.WatchConfig{
		FilePath:    path,
		Environment: "test",
		Interval:    time.Second,
	})

	// Baseline poll
	_, _ = w.Poll()

	// Mutate the file
	if err := os.WriteFile(path, []byte("KEY=changed\n"), 0644); err != nil {
		t.Fatalf("failed to update file: %v", err)
	}

	event, err := w.Poll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected drift event, got nil")
	}
	if len(event.Changes) == 0 {
		t.Error("expected at least one change in drift event")
	}
}

func TestWatcher_DriftCount(t *testing.T) {
	path := writeTempConfig(t, "A=1\n")
	w := watch.New(watch.WatchConfig{FilePath: path, Environment: "test"})
	_, _ = w.Poll()

	for i, content := range []string{"A=2\n", "A=3\n"} {
		_ = os.WriteFile(path, []byte(content), 0644)
		_, err := w.Poll()
		if err != nil {
			t.Fatalf("poll %d failed: %v", i, err)
		}
	}

	if got := w.DriftCount(); got != 2 {
		t.Errorf("expected drift count 2, got %d", got)
	}
}

func TestWatcher_MissingFile(t *testing.T) {
	w := watch.New(watch.WatchConfig{
		FilePath:    "/nonexistent/path/config.env",
		Environment: "test",
	})
	_, err := w.Poll()
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
