package annotate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/patchwork/internal/annotate"
)

func emptyStore() *annotate.Store {
	return &annotate.Store{Annotations: make(map[string]annotate.Annotation)}
}

func TestAdd_CreatesAnnotation(t *testing.T) {
	store := emptyStore()
	a := annotate.Add(store, "entry-1", "initial deploy")
	if a.EntryID != "entry-1" {
		t.Errorf("expected entry-1, got %s", a.EntryID)
	}
	if a.Note != "initial deploy" {
		t.Errorf("unexpected note: %s", a.Note)
	}
	if a.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestAdd_OverwritesExisting(t *testing.T) {
	store := emptyStore()
	annotate.Add(store, "entry-1", "first note")
	annotate.Add(store, "entry-1", "updated note")
	a, ok := annotate.Get(store, "entry-1")
	if !ok {
		t.Fatal("expected annotation to exist")
	}
	if a.Note != "updated note" {
		t.Errorf("expected updated note, got %s", a.Note)
	}
}

func TestRemove_ExistingAnnotation(t *testing.T) {
	store := emptyStore()
	annotate.Add(store, "entry-2", "to be removed")
	if err := annotate.Remove(store, "entry-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := annotate.Get(store, "entry-2")
	if ok {
		t.Error("expected annotation to be removed")
	}
}

func TestRemove_MissingAnnotation(t *testing.T) {
	store := emptyStore()
	err := annotate.Remove(store, "nonexistent")
	if err == nil {
		t.Error("expected error for missing annotation")
	}
}

func TestGet_MissingAnnotation(t *testing.T) {
	store := emptyStore()
	_, ok := annotate.Get(store, "missing")
	if ok {
		t.Error("expected Get to return false for missing entry")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "annotations.json")

	store := emptyStore()
	annotate.Add(store, "entry-3", "round trip note")

	if err := annotate.Save(store, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := annotate.Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	a, ok := annotate.Get(loaded, "entry-3")
	if !ok {
		t.Fatal("expected annotation after load")
	}
	if a.Note != "round trip note" {
		t.Errorf("unexpected note: %s", a.Note)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	store, err := annotate.Load(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if store == nil || store.Annotations == nil {
		t.Error("expected non-nil empty store")
	}
	_ = os.Remove(path) // no-op, file never created
}
