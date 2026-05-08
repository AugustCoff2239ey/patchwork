package annotate

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Annotation holds a user-defined note attached to a history entry by ID.
type Annotation struct {
	EntryID   string    `json:"entry_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// Store is a collection of annotations keyed by entry ID.
type Store struct {
	Annotations map[string]Annotation `json:"annotations"`
}

// Add attaches a note to the given entry ID. Overwrites any existing annotation.
func Add(store *Store, entryID, note string) Annotation {
	if store.Annotations == nil {
		store.Annotations = make(map[string]Annotation)
	}
	a := Annotation{
		EntryID:   entryID,
		Note:      note,
		CreatedAt: time.Now().UTC(),
	}
	store.Annotations[entryID] = a
	return a
}

// Remove deletes the annotation for the given entry ID.
// Returns an error if no annotation exists.
func Remove(store *Store, entryID string) error {
	if store.Annotations == nil {
		return fmt.Errorf("no annotation found for entry %q", entryID)
	}
	if _, ok := store.Annotations[entryID]; !ok {
		return fmt.Errorf("no annotation found for entry %q", entryID)
	}
	delete(store.Annotations, entryID)
	return nil
}

// Get retrieves the annotation for the given entry ID.
func Get(store *Store, entryID string) (Annotation, bool) {
	if store.Annotations == nil {
		return Annotation{}, false
	}
	a, ok := store.Annotations[entryID]
	return a, ok
}

// Save writes the annotation store to the given path as JSON.
func Save(store *Store, path string) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("annotate: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("annotate: write: %w", err)
	}
	return nil
}

// Load reads an annotation store from the given path.
// Returns an empty store if the file does not exist.
func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{Annotations: make(map[string]Annotation)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("annotate: read: %w", err)
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("annotate: unmarshal: %w", err)
	}
	if store.Annotations == nil {
		store.Annotations = make(map[string]Annotation)
	}
	return &store, nil
}
