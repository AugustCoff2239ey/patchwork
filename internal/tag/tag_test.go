package tag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_NewTag(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	if len(tags["snap-1"]) != 1 || tags["snap-1"][0] != "production" {
		t.Fatalf("expected tag to be added, got %v", tags["snap-1"])
	}
}

func TestAdd_DuplicateTag(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	tags = Add(tags, "snap-1", "production")
	if len(tags["snap-1"]) != 1 {
		t.Fatalf("expected no duplicate, got %v", tags["snap-1"])
	}
}

func TestAdd_MultipleTags_Sorted(t *testing.T) {
	tags := Add(nil, "snap-1", "zebra")
	tags = Add(tags, "snap-1", "alpha")
	if tags["snap-1"][0] != "alpha" || tags["snap-1"][1] != "zebra" {
		t.Fatalf("expected sorted tags, got %v", tags["snap-1"])
	}
}

func TestRemove_ExistingTag(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	tags, err := Remove(tags, "snap-1", "production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags["snap-1"]) != 0 {
		t.Fatalf("expected tag to be removed, got %v", tags["snap-1"])
	}
}

func TestRemove_MissingTag(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	_, err := Remove(tags, "snap-1", "staging")
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestRemove_MissingSnapshot(t *testing.T) {
	_, err := Remove(nil, "snap-999", "production")
	if err == nil {
		t.Fatal("expected error for nil tags")
	}
}

func TestFindByTag(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	tags = Add(tags, "snap-2", "production")
	tags = Add(tags, "snap-3", "staging")

	result := FindByTag(tags, "production")
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "snap-1" || result[1] != "snap-2" {
		t.Fatalf("unexpected results: %v", result)
	}
}

func TestFindByTag_NotFound(t *testing.T) {
	tags := Add(nil, "snap-1", "production")
	result := FindByTag(tags, "staging")
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.json")

	tags := Add(nil, "snap-1", "production")
	tags = Add(tags, "snap-2", "staging")

	if err := Save(path, tags); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded["snap-1"]) != 1 || loaded["snap-1"][0] != "production" {
		t.Fatalf("unexpected loaded tags: %v", loaded)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	tags, err := Load("/nonexistent/path/tags.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if tags == nil || len(tags) != 0 {
		t.Fatalf("expected empty TagMap, got %v", tags)
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.json")
	if err := Save(path, make(TagMap)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to be created")
	}
}
