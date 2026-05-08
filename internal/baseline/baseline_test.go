package baseline

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

func makeEntry(service string, fields map[string]any) parser.Entry {
	return parser.Entry{
		Service: service,
		Fields:  fields,
	}
}

func TestNew_InvalidTTL_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
	_, err = New(-time.Second)
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestNew_ValidTTL_ReturnsBaseline(t *testing.T) {
	b, err := New(time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil baseline")
	}
}

func TestCompare_NoBaseline_ReturnsEmpty(t *testing.T) {
	b, _ := New(time.Minute)
	entry := makeEntry("svcA", map[string]any{"level": "info"})
	devs := b.Compare(entry)
	if len(devs) != 0 {
		t.Fatalf("expected no deviations without baseline, got %d", len(devs))
	}
}

func TestCompare_IdenticalEntry_NoDeviations(t *testing.T) {
	b, _ := New(time.Minute)
	entry := makeEntry("svcA", map[string]any{"level": "info", "env": "prod"})
	b.Record(entry)
	devs := b.Compare(entry)
	if len(devs) != 0 {
		t.Fatalf("expected no deviations, got %v", devs)
	}
}

func TestCompare_ChangedField_ReturnsDeviation(t *testing.T) {
	b, _ := New(time.Minute)
	base := makeEntry("svcA", map[string]any{"level": "info"})
	b.Record(base)

	newer := makeEntry("svcA", map[string]any{"level": "error"})
	devs := b.Compare(newer)
	if len(devs) != 1 {
		t.Fatalf("expected 1 deviation, got %d", len(devs))
	}
	if devs[0].Field != "level" || devs[0].Expected != "info" || devs[0].Actual != "error" {
		t.Errorf("unexpected deviation: %+v", devs[0])
	}
}

func TestCompare_MissingField_ReturnsDeviation(t *testing.T) {
	b, _ := New(time.Minute)
	b.Record(makeEntry("svcA", map[string]any{"level": "info", "env": "prod"}))
	devs := b.Compare(makeEntry("svcA", map[string]any{"level": "info"}))
	if len(devs) != 1 || devs[0].Actual != "<missing>" {
		t.Errorf("expected missing-field deviation, got %v", devs)
	}
}

func TestRecord_ExpiredTTL_RefreshesBaseline(t *testing.T) {
	b, _ := New(10 * time.Millisecond)
	b.Record(makeEntry("svcA", map[string]any{"level": "info"}))
	time.Sleep(20 * time.Millisecond)
	// After TTL expires, a new Record should overwrite the old baseline.
	b.Record(makeEntry("svcA", map[string]any{"level": "warn"}))
	devs := b.Compare(makeEntry("svcA", map[string]any{"level": "warn"}))
	if len(devs) != 0 {
		t.Errorf("expected no deviations after refresh, got %v", devs)
	}
}

func TestReset_ClearsBaseline(t *testing.T) {
	b, _ := New(time.Minute)
	b.Record(makeEntry("svcA", map[string]any{"level": "info"}))
	b.Reset("svcA")
	devs := b.Compare(makeEntry("svcA", map[string]any{"level": "error"}))
	if len(devs) != 0 {
		t.Errorf("expected no deviations after reset, got %v", devs)
	}
}
