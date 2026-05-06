package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork/internal/snapshot"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cfg-*.conf")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestCapture_ReturnsSnapshot(t *testing.T) {
	const content = "key=value\nfoo=bar\n"
	path := writeTempFile(t, content)

	s, err := snapshot.Capture(path)
	if err != nil {
		t.Fatalf("Capture() error: %v", err)
	}

	if s.Content != content {
		t.Errorf("Content = %q, want %q", s.Content, content)
	}
	if s.Checksum == "" {
		t.Error("Checksum should not be empty")
	}
	if s.ID == "" {
		t.Error("ID should not be empty")
	}
	if s.FilePath != path {
		t.Errorf("FilePath = %q, want %q", s.FilePath, path)
	}
}

func TestCapture_MissingFile(t *testing.T) {
	_, err := snapshot.Capture("/nonexistent/path/config.conf")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := writeTempFile(t, "timeout=30\n")
	s, err := snapshot.Capture(path)
	if err != nil {
		t.Fatalf("Capture() error: %v", err)
	}

	dest := filepath.Join(t.TempDir(), "snap.json")
	if err := s.Save(dest); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := snapshot.Load(dest)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.ID != s.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, s.ID)
	}
	if loaded.Checksum != s.Checksum {
		t.Errorf("Checksum mismatch: got %q, want %q", loaded.Checksum, s.Checksum)
	}
	if loaded.Content != s.Content {
		t.Errorf("Content mismatch: got %q, want %q", loaded.Content, s.Content)
	}
}
