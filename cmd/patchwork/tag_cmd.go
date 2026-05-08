package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/patchwork/internal/tag"
)

const defaultTagFile = ".patchwork_tags.json"

func runTagAdd(snapshotID, label, tagFile string) error {
	if snapshotID == "" || label == "" {
		return fmt.Errorf("tag add requires a snapshot ID and a tag label")
	}
	tags, err := tag.Load(tagFile)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	tags = tag.Add(tags, snapshotID, label)
	if err := tag.Save(tagFile, tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}
	fmt.Fprintf(os.Stdout, "tagged snapshot %q with %q\n", snapshotID, label)
	return nil
}

func runTagRemove(snapshotID, label, tagFile string) error {
	if snapshotID == "" || label == "" {
		return fmt.Errorf("tag remove requires a snapshot ID and a tag label")
	}
	tags, err := tag.Load(tagFile)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	tags, err = tag.Remove(tags, snapshotID, label)
	if err != nil {
		return fmt.Errorf("tag remove: %w", err)
	}
	if err := tag.Save(tagFile, tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}
	fmt.Fprintf(os.Stdout, "removed tag %q from snapshot %q\n", label, snapshotID)
	return nil
}

func runTagList(snapshotID, tagFile string) error {
	tags, err := tag.Load(tagFile)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	if snapshotID != "" {
		list := tags[snapshotID]
		if len(list) == 0 {
			fmt.Fprintf(os.Stdout, "no tags for snapshot %q\n", snapshotID)
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", snapshotID, strings.Join(list, ", "))
		return nil
	}
	if len(tags) == 0 {
		fmt.Fprintln(os.Stdout, "no tags found")
		return nil
	}
	for id, list := range tags {
		fmt.Fprintf(os.Stdout, "%s: %s\n", id, strings.Join(list, ", "))
	}
	return nil
}

func runTagFind(label, tagFile string) error {
	if label == "" {
		return fmt.Errorf("tag find requires a tag label")
	}
	tags, err := tag.Load(tagFile)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	results := tag.FindByTag(tags, label)
	if len(results) == 0 {
		fmt.Fprintf(os.Stdout, "no snapshots found with tag %q\n", label)
		return nil
	}
	for _, id := range results {
		fmt.Fprintln(os.Stdout, id)
	}
	return nil
}
