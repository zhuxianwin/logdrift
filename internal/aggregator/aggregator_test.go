package aggregator

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/parser"
)

func makeEntry(service, keyVal string, ts time.Time) parser.Entry {
	return parser.Entry{
		Service:   service,
		Timestamp: ts,
		Fields:    map[string]interface{}{"level": keyVal, "msg": "hello"},
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := New("", time.Minute)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := New("level", 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNew_Valid_ReturnsAggregator(t *testing.T) {
	a, err := New("level", time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil aggregator")
	}
}

func TestAdd_GroupsEntriesByField(t *testing.T) {
	a, _ := New("level", time.Minute)
	now := time.Now()
	a.Add(makeEntry("svc-a", "error", now))
	a.Add(makeEntry("svc-b", "error", now.Add(time.Second)))
	a.Add(makeEntry("svc-a", "info", now.Add(2*time.Second)))

	groups := a.Groups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	for _, g := range groups {
		switch g.Key {
		case "error":
			if g.Count != 2 {
				t.Errorf("error group count: want 2, got %d", g.Count)
			}
		case "info":
			if g.Count != 1 {
				t.Errorf("info group count: want 1, got %d", g.Count)
			}
		}
	}
}

func TestAdd_MissingField_Ignored(t *testing.T) {
	a, _ := New("level", time.Minute)
	entry := parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Fields:    map[string]interface{}{"msg": "no level here"},
	}
	a.Add(entry)
	if len(a.Groups()) != 0 {
		t.Fatal("expected no groups when field is absent")
	}
}

func TestGroups_EvictsStaleEntries(t *testing.T) {
	a, _ := New("level", 5*time.Second)
	past := time.Now().Add(-10 * time.Second)
	a.Add(makeEntry("svc", "warn", past))
	// groups should be evicted since LastSeen is outside window
	if len(a.Groups()) != 0 {
		t.Fatal("expected stale group to be evicted")
	}
}

func TestFromConfig_Nil_ReturnsNil(t *testing.T) {
	agg, err := FromConfig(nil)
	if err != nil || agg != nil {
		t.Fatalf("expected nil, nil for nil config; got %v, %v", agg, err)
	}
}

func TestFromConfig_ValidConfig_ReturnsAggregator(t *testing.T) {
	cfg := &Config{Field: "level", Window: "30s"}
	agg, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agg == nil {
		t.Fatal("expected non-nil aggregator")
	}
}

func TestFromConfig_InvalidWindow_ReturnsError(t *testing.T) {
	cfg := &Config{Field: "level", Window: "not-a-duration"}
	_, err := FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid window")
	}
}
