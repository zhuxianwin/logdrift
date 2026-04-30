package formatter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/differ"
	"github.com/yourorg/logdrift/internal/formatter"
)

var fixedTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func deltas() []differ.Delta {
	return []differ.Delta{
		{Type: differ.DeltaChanged, Field: "level", ValueA: "info", ValueB: "error"},
		{Type: differ.DeltaAdded, Field: "trace_id", ValueB: "abc123"},
		{Type: differ.DeltaRemoved, Field: "request_id", ValueA: "xyz"},
	}
}

func TestFormatPlain_ReturnsNilForNoDeltas(t *testing.T) {
	lines := formatter.FormatPlain("svcA", "svcB", fixedTime, nil)
	if lines != nil {
		t.Fatalf("expected nil for empty deltas, got %v", lines)
	}
}

func TestFormatPlain_HeaderIncludesServices(t *testing.T) {
	lines := formatter.FormatPlain("svcA", "svcB", fixedTime, deltas())
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	header := lines[0]
	if !strings.Contains(header, "svcA") || !strings.Contains(header, "svcB") {
		t.Errorf("header missing service names: %q", header)
	}
	if !strings.Contains(header, "2024-06-01") {
		t.Errorf("header missing timestamp: %q", header)
	}
}

func TestFormatPlain_LineCountMatchesDeltas(t *testing.T) {
	d := deltas()
	lines := formatter.FormatPlain("svcA", "svcB", fixedTime, d)
	// header + one line per delta
	if len(lines) != len(d)+1 {
		t.Errorf("expected %d lines, got %d", len(d)+1, len(lines))
	}
}

func TestFormatPlain_ChangedPrefix(t *testing.T) {
	lines := formatter.FormatPlain("a", "b", fixedTime, deltas())
	changedLine := lines[1] // first delta is Changed
	if !strings.Contains(changedLine, "~") {
		t.Errorf("changed delta should contain '~', got %q", changedLine)
	}
	if !strings.Contains(changedLine, "info") || !strings.Contains(changedLine, "error") {
		t.Errorf("changed line should show both values: %q", changedLine)
	}
}

func TestFormatPlain_AddedPrefix(t *testing.T) {
	lines := formatter.FormatPlain("a", "b", fixedTime, deltas())
	addedLine := lines[2]
	if !strings.HasPrefix(strings.TrimSpace(addedLine), "+") {
		t.Errorf("added delta should start with '+', got %q", addedLine)
	}
}

func TestFormatPlain_RemovedPrefix(t *testing.T) {
	lines := formatter.FormatPlain("a", "b", fixedTime, deltas())
	removedLine := lines[3]
	if !strings.HasPrefix(strings.TrimSpace(removedLine), "-") {
		t.Errorf("removed delta should start with '-', got %q", removedLine)
	}
}

func TestJoin_ConcatenatesWithNewlines(t *testing.T) {
	lines := []string{"line1", "line2", "line3"}
	result := formatter.Join(lines)
	if result != "line1\nline2\nline3" {
		t.Errorf("unexpected join result: %q", result)
	}
}
