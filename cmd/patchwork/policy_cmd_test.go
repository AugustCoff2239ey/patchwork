package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork/internal/history"
	"github.com/patchwork/internal/policy"
	"github.com/patchwork/internal/snapshot"
)

func writePolicyFile(t *testing.T, pf policy.PolicyFile) string {
	t.Helper()
	data, _ := json.Marshal(pf)
	path := filepath.Join(t.TempDir(), "policy.json")
	os.WriteFile(path, data, 0644)
	return path
}

func writePolicyHistory(t *testing.T, dir string, entries []history.Entry) string {
	t.Helper()
	log := history.Log{Entries: entries}
	data, _ := json.Marshal(log)
	path := filepath.Join(dir, "history.json")
	os.WriteFile(path, data, 0644)
	return path
}

func writePolicySnap(t *testing.T, dir, name string, data map[string]string) string {
	t.Helper()
	snap := snapshot.Snapshot{Data: data}
	path := filepath.Join(dir, name)
	snapshot.Save(snap, path)
	return path
}

func TestRunPolicy_MissingPolicyFlag(t *testing.T) {
	err := runPolicy("", "history.json", "prod")
	if err == nil {
		t.Error("expected error for missing policy flag")
	}
}

func TestRunPolicy_MissingEnvFlag(t *testing.T) {
	err := runPolicy("policy.json", "history.json", "")
	if err == nil {
		t.Error("expected error for missing env flag")
	}
}

func TestRunPolicy_MissingHistoryFile(t *testing.T) {
	dir := t.TempDir()
	pf := policy.PolicyFile{Rules: []policy.Rule{{Name: "r", MaxChanges: 5}}}
	pPath := writePolicyFile(t, pf)
	err := runPolicy(pPath, filepath.Join(dir, "missing.json"), "prod")
	if err == nil {
		t.Error("expected error for missing history file")
	}
}

func TestRunPolicy_NotEnoughEntries(t *testing.T) {
	dir := t.TempDir()
	pf := policy.PolicyFile{Rules: []policy.Rule{{Name: "r", MaxChanges: 5}}}
	pPath := writePolicyFile(t, pf)
	s1 := writePolicySnap(t, dir, "s1.json", map[string]string{"A": "1"})
	entries := []history.Entry{{Environment: "prod", SnapshotPath: s1, Timestamp: time.Now()}}
	hPath := writePolicyHistory(t, dir, entries)
	err := runPolicy(pPath, hPath, "prod")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunPolicy_NoViolations(t *testing.T) {
	dir := t.TempDir()
	pf := policy.PolicyFile{Rules: []policy.Rule{{Name: "r", MaxChanges: 10}}}
	pPath := writePolicyFile(t, pf)
	s1 := writePolicySnap(t, dir, "s1.json", map[string]string{"A": "1"})
	s2 := writePolicySnap(t, dir, "s2.json", map[string]string{"A": "2"})
	entries := []history.Entry{
		{Environment: "prod", SnapshotPath: s1, Timestamp: time.Now().Add(-time.Hour)},
		{Environment: "prod", SnapshotPath: s2, Timestamp: time.Now()},
	}
	hPath := writePolicyHistory(t, dir, entries)
	err := runPolicy(pPath, hPath, "prod")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
