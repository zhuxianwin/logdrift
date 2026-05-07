package correlation

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/parser"
)

func makeEntry(id string, ts time.Time) parser.Entry {
	fields := map[string]interface{}{"msg": "hello"}
	if id != "" {
		fields["trace_id"] = id
	}
	return parser.Entry{Timestamp: ts, Message: "hello", Fields: fields}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := New("", time.Minute)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ZeroTTL_ReturnsError(t *testing.T) {
	_, err := New("trace_id", 0)
	if err == nil {
		t.Fatal("expected error for zero ttl")
	}
}

func TestAdd_NoCorrelationField_ReturnsEmpty(t *testing.T) {
	tr, _ := New("trace_id", time.Minute)
	e := makeEntry("", time.Now())
	if id := tr.Add("svc", e); id != "" {
		t.Fatalf("expected empty id, got %q", id)
	}
	if tr.Len() != 0 {
		t.Fatal("expected no groups")
	}
}

func TestAdd_CreatesGroup(t *testing.T) {
	tr, _ := New("trace_id", time.Minute)
	now := time.Now()
	tr.Add("svc-a", makeEntry("req-1", now))
	tr.Add("svc-b", makeEntry("req-1", now.Add(time.Millisecond)))

	if tr.Len() != 1 {
		t.Fatalf("expected 1 group, got %d", tr.Len())
	}
	g := tr.Get("req-1")
	if g == nil {
		t.Fatal("expected group for req-1")
	}
	if len(g.Entries["svc-a"]) != 1 || len(g.Entries["svc-b"]) != 1 {
		t.Fatal("expected one entry per service")
	}
}

func TestAdd_MultipleIDs_CreatesMultipleGroups(t *testing.T) {
	tr, _ := New("trace_id", time.Minute)
	now := time.Now()
	tr.Add("svc", makeEntry("req-1", now))
	tr.Add("svc", makeEntry("req-2", now))
	if tr.Len() != 2 {
		t.Fatalf("expected 2 groups, got %d", tr.Len())
	}
}

func TestEvict_RemovesExpiredGroups(t *testing.T) {
	tr, _ := New("trace_id", time.Minute)
	old := time.Now().Add(-2 * time.Minute)
	recent := time.Now()
	tr.Add("svc", makeEntry("old-req", old))
	tr.Add("svc", makeEntry("new-req", recent))

	removed := tr.Evict(time.Now())
	if removed != 1 {
		t.Fatalf("expected 1 eviction, got %d", removed)
	}
	if tr.Len() != 1 {
		t.Fatalf("expected 1 remaining group, got %d", tr.Len())
	}
	if tr.Get("old-req") != nil {
		t.Fatal("expected old-req to be evicted")
	}
}

func TestGet_UnknownID_ReturnsNil(t *testing.T) {
	tr, _ := New("trace_id", time.Minute)
	if g := tr.Get("nope"); g != nil {
		t.Fatal("expected nil for unknown id")
	}
}
