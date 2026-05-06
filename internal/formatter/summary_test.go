package formatter

import (
	"strings"
	"testing"

	"github.com/user/logdrift/internal/differ"
)

func makeDelta(field, a, b string) differ.Delta {
	return differ.Delta{Field: field, A: a, B: b}
}

func TestSummaryStats_EmptySession(t *testing.T) {
	s := NewSummaryStats([]string{"svc-a", "svc-b"})
	out := s.Render(NoColorScheme())

	if s.TotalLines != 0 || s.Divergences != 0 {
		t.Fatalf("expected zero counts, got lines=%d divergences=%d", s.TotalLines, s.Divergences)
	}
	if !strings.Contains(out, "svc-a") {
		t.Error("expected service name in summary output")
	}
}

func TestSummaryStats_RecordNoDelta(t *testing.T) {
	s := NewSummaryStats([]string{"svc-a"})
	s.Record(nil)
	s.Record([]differ.Delta{})

	if s.TotalLines != 2 {
		t.Fatalf("expected TotalLines=2, got %d", s.TotalLines)
	}
	if s.Divergences != 0 {
		t.Fatalf("expected Divergences=0, got %d", s.Divergences)
	}
}

func TestSummaryStats_RecordDeltas(t *testing.T) {
	s := NewSummaryStats([]string{"svc-a", "svc-b"})
	s.Record([]differ.Delta{makeDelta("level", "info", "error")})
	s.Record([]differ.Delta{makeDelta("level", "warn", "error"), makeDelta("msg", "x", "y")})
	s.Record(nil)

	if s.TotalLines != 3 {
		t.Fatalf("expected TotalLines=3, got %d", s.TotalLines)
	}
	if s.Divergences != 2 {
		t.Fatalf("expected Divergences=2, got %d", s.Divergences)
	}
	if s.FieldChanges["level"] != 2 {
		t.Errorf("expected level count=2, got %d", s.FieldChanges["level"])
	}
	if s.FieldChanges["msg"] != 1 {
		t.Errorf("expected msg count=1, got %d", s.FieldChanges["msg"])
	}
}

func TestSummaryStats_RenderContainsDriftPercent(t *testing.T) {
	s := NewSummaryStats([]string{"a", "b"})
	for i := 0; i < 4; i++ {
		s.Record(nil)
	}
	s.Record([]differ.Delta{makeDelta("status", "200", "500")})

	out := s.Render(NoColorScheme())
	if !strings.Contains(out, "20.0%") {
		t.Errorf("expected 20.0%% drift in output, got:\n%s", out)
	}
}

func TestSummaryStats_RenderTopFields(t *testing.T) {
	s := NewSummaryStats([]string{"a"})
	s.Record([]differ.Delta{makeDelta("latency", "10", "20")})

	out := s.Render(NoColorScheme())
	if !strings.Contains(out, "latency") {
		t.Errorf("expected field 'latency' in summary output, got:\n%s", out)
	}
}
