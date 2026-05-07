package schema

import (
	"testing"
	"time"

	"github.com/robinmuhia/logdrift/internal/parser"
)

func makeEntry(raw map[string]any) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Raw:       raw,
	}
}

func TestNew_EmptySchema(t *testing.T) {
	s := New()
	if s.Total() != 0 {
		t.Fatalf("expected 0 total, got %d", s.Total())
	}
	if len(s.Fields()) != 0 {
		t.Fatalf("expected no fields, got %v", s.Fields())
	}
}

func TestObserve_RecordsFields(t *testing.T) {
	s := New()
	s.Observe(makeEntry(map[string]any{"level": "info", "msg": "hello"}))

	fields := s.Fields()
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0] != "level" || fields[1] != "msg" {
		t.Errorf("unexpected fields: %v", fields)
	}
	if s.Total() != 1 {
		t.Fatalf("expected total 1, got %d", s.Total())
	}
}

func TestObserve_AccumulatesValueCounts(t *testing.T) {
	s := New()
	s.Observe(makeEntry(map[string]any{"level": "info"}))
	s.Observe(makeEntry(map[string]any{"level": "info"}))
	s.Observe(makeEntry(map[string]any{"level": "error"}))

	fs, ok := s.Stats("level")
	if !ok {
		t.Fatal("expected field 'level' to exist")
	}
	if fs.Count != 3 {
		t.Errorf("expected count 3, got %d", fs.Count)
	}
	if fs.Values["info"] != 2 {
		t.Errorf("expected info=2, got %d", fs.Values["info"])
	}
	if fs.Values["error"] != 1 {
		t.Errorf("expected error=1, got %d", fs.Values["error"])
	}
}

func TestCoverage_PartialPresence(t *testing.T) {
	s := New()
	s.Observe(makeEntry(map[string]any{"level": "info", "trace": "abc"}))
	s.Observe(makeEntry(map[string]any{"level": "warn"}))

	if got := s.Coverage("level"); got != 1.0 {
		t.Errorf("expected coverage 1.0 for level, got %f", got)
	}
	if got := s.Coverage("trace"); got != 0.5 {
		t.Errorf("expected coverage 0.5 for trace, got %f", got)
	}
}

func TestCoverage_MissingField_ReturnsZero(t *testing.T) {
	s := New()
	s.Observe(makeEntry(map[string]any{"level": "info"}))

	if got := s.Coverage("nonexistent"); got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestCoverage_ZeroTotal_ReturnsZero(t *testing.T) {
	s := New()
	if got := s.Coverage("level"); got != 0 {
		t.Errorf("expected 0 for empty schema, got %f", got)
	}
}

func TestStats_MissingField_ReturnsFalse(t *testing.T) {
	s := New()
	_, ok := s.Stats("ghost")
	if ok {
		t.Error("expected false for missing field")
	}
}
