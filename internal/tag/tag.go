package tag

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// TagMap holds a mapping of snapshot IDs to user-defined tags.
type TagMap map[string][]string

// Add associates a tag with a snapshot ID.
func Add(tags TagMap, snapshotID, tag string) TagMap {
	if tags == nil {
		tags = make(TagMap)
	}
	for _, existing := range tags[snapshotID] {
		if existing == tag {
			return tags
		}
	}
	tags[snapshotID] = append(tags[snapshotID], tag)
	sort.Strings(tags[snapshotID])
	return tags
}

// Remove removes a tag from a snapshot ID. Returns an error if the tag is not found.
func Remove(tags TagMap, snapshotID, tag string) (TagMap, error) {
	if tags == nil {
		return nil, fmt.Errorf("tag %q not found for snapshot %q", tag, snapshotID)
	}
	list, ok := tags[snapshotID]
	if !ok {
		return tags, fmt.Errorf("tag %q not found for snapshot %q", tag, snapshotID)
	}
	updated := list[:0]
	found := false
	for _, t := range list {
		if t == tag {
			found = true
			continue
		}
		updated = append(updated, t)
	}
	if !found {
		return tags, fmt.Errorf("tag %q not found for snapshot %q", tag, snapshotID)
	}
	tags[snapshotID] = updated
	return tags, nil
}

// FindByTag returns all snapshot IDs that have the given tag.
func FindByTag(tags TagMap, tag string) []string {
	var result []string
	for id, list := range tags {
		for _, t := range list {
			if t == tag {
				result = append(result, id)
				break
			}
		}
	}
	sort.Strings(result)
	return result
}

// Save writes the TagMap to a JSON file at the given path.
func Save(path string, tags TagMap) error {
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("tag: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a TagMap from a JSON file. Returns an empty map if the file does not exist.
func Load(path string) (TagMap, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(TagMap), nil
	}
	if err != nil {
		return nil, fmt.Errorf("tag: read: %w", err)
	}
	var tags TagMap
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("tag: unmarshal: %w", err)
	}
	return tags, nil
}
