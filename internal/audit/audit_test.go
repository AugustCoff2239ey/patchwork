package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func emptyLog() *Log {
	return &Log{}
}

func TestAppend_AddsEvent(t *testing.T) {
	log := emptyLog()
	Append(log, EventCapture, "production", "captured config")
	if len(log.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(log.Events))
	}
	if log.Events[0].Kind != EventCapture {
		t.Errorf("expected kind %q, got %q", EventCapture, log.Events[0].Kind)
	}
	if log.Events[0].Environment != "production" {
		t.Errorf("expected env production, got %q", log.Events[0].Environment)
	}
}

func TestFilter_ByKind(t *testing.T) {
	log := emptyLog()
	Append(log, EventCapture, "staging", "cap")
	Append(log, EventDiff, "staging", "diff")
	Append(log, EventCapture, "production", "cap")

	results := Filter(log, EventCapture, "")
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestFilter_ByEnvironment(t *testing.T) {
	log := emptyLog()
	Append(log, EventCapture, "staging", "cap")
	Append(log, EventDiff, "production", "diff")

	results := Filter(log, "", "staging")
	if len(results) != 1 {
		t.Errorf("expected 1, got %d", len(results))
	}
	if results[0].Environment != "staging" {
		t.Errorf("unexpected env %q", results[0].Environment)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	log := emptyLog()
	Append(log, EventCapture, "staging", "cap")

	results := Filter(log, EventRollback, "")
	if len(results) != 0 {
		t.Errorf("expected 0, got %d", len(results))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	log := emptyLog()
	Append(log, EventExport, "dev", "exported report")

	if err := Save(path, log); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(loaded.Events))
	}
	if loaded.Events[0].Kind != EventExport {
		t.Errorf("expected kind export, got %q", loaded.Events[0].Kind)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	log, err := Load("/nonexistent/audit.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(log.Events) != 0 {
		t.Errorf("expected empty log, got %d events", len(log.Events))
	}
}

func TestRender_EmptyEvents(t *testing.T) {
	out := Render([]Event{})
	if out != "No audit events found.\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRender_ContainsKindAndEnv(t *testing.T) {
	events := []Event{
		{Timestamp: time.Now().UTC(), Kind: EventBaseline, Environment: "staging", Message: "pinned", User: "alice"},
	}
	out := Render(events)
	if !contains(out, "baseline") {
		t.Errorf("expected kind in output: %q", out)
	}
	if !contains(out, "staging") {
		t.Errorf("expected env in output: %q", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestSave_SortsEventsByTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	log := emptyLog()
	later := time.Now().UTC()
	earlier := later.Add(-time.Hour)

	log.Events = append(log.Events,
		Event{Timestamp: later, Kind: EventDiff, Environment: "prod", Message: "b"},
		Event{Timestamp: earlier, Kind: EventCapture, Environment: "prod", Message: "a"},
	)

	if err := Save(path, log); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, _ := Load(path)
	if loaded.Events[0].Kind != EventCapture {
		t.Errorf("expected first event to be capture (earlier), got %q", loaded.Events[0].Kind)
	}
	_ = os.Remove(path)
}
