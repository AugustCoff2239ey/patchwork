package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patchwork/internal/alert"
	"github.com/patchwork/internal/diff"
	"github.com/patchwork/internal/history"
)

// Format represents the output format for an export.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// ExportRecord holds all data for a single exported snapshot comparison.
type ExportRecord struct {
	Timestamp   time.Time      `json:"timestamp"`
	Environment string         `json:"environment"`
	Changes     []diff.Change  `json:"changes"`
	Alerts      []alert.Alert  `json:"alerts"`
	Summary     Summary        `json:"summary"`
}

// Summary provides a high-level count of changes by type.
type Summary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Total    int `json:"total"`
}

// Build constructs an ExportRecord from a history entry and its associated changes.
func Build(entry history.Entry, changes []diff.Change, alerts []alert.Alert) ExportRecord {
	s := Summary{Total: len(changes)}
	for _, c := range changes {
		switch c.Type {
		case diff.Added:
			s.Added++
		case diff.Removed:
			s.Removed++
		case diff.Modified:
			s.Modified++
		}
	}
	return ExportRecord{
		Timestamp:   entry.Timestamp,
		Environment: entry.Environment,
		Changes:     changes,
		Alerts:      alerts,
		Summary:     s,
	}
}

// Write serialises the ExportRecord to a file in the given format.
func Write(rec ExportRecord, dir string, format Format) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("export: create dir: %w", err)
	}

	filename := fmt.Sprintf("export_%s_%s.%s",
		rec.Environment,
		rec.Timestamp.Format("20060102T150405"),
		string(format),
	)
	path := filepath.Join(dir, filename)

	switch format {
	case FormatJSON:
		data, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			return "", fmt.Errorf("export: marshal: %w", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return "", fmt.Errorf("export: write file: %w", err)
		}
	case FormatText:
		f, err := os.Create(path)
		if err != nil {
			return "", fmt.Errorf("export: create file: %w", err)
		}
		defer f.Close()
		fmt.Fprintf(f, "Patchwork Export\n")
		fmt.Fprintf(f, "Environment : %s\n", rec.Environment)
		fmt.Fprintf(f, "Timestamp   : %s\n", rec.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(f, "Changes     : +%d -%d ~%d (total %d)\n",
			rec.Summary.Added, rec.Summary.Removed, rec.Summary.Modified, rec.Summary.Total)
		fmt.Fprintf(f, "Alerts      : %d\n\n", len(rec.Alerts))
		for _, c := range rec.Changes {
			fmt.Fprintf(f, "  [%s] %s\n", c.Type, c.Key)
		}
	default:
		return "", fmt.Errorf("export: unsupported format %q", format)
	}

	return path, nil
}
