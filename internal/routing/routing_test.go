package routing

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/parser"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Fields:    fields,
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New([]Rule{{Field: "level", Pattern: "[", Channel: "errors"}})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := New([]Rule{{Field: "", Pattern: ".*", Channel: "all"}})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_EmptyChannel_ReturnsError(t *testing.T) {
	_, err := New([]Rule{{Field: "level", Pattern: ".*", Channel: ""}})
	if err == nil {
		t.Fatal("expected error for empty channel")
	}
}

func TestRoute_MatchingSendsToChannel(t *testing.T) {
	rt, err := New([]Rule{{Field: "level", Pattern: "^error$", Channel: "errors"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch := rt.Channel("errors")
	entry := makeEntry(map[string]any{"level": "error", "msg": "boom"})
	rt.Route(entry)

	select {
	case got := <-ch:
		if got.Fields["msg"] != "boom" {
			t.Errorf("unexpected entry: %v", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for routed entry")
	}
}

func TestRoute_NoMatch_DoesNotSend(t *testing.T) {
	rt, err := New([]Rule{{Field: "level", Pattern: "^error$", Channel: "errors"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch := rt.Channel("errors")
	rt.Route(makeEntry(map[string]any{"level": "info"}))

	select {
	case <-ch:
		t.Fatal("expected no entry to be routed")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRoute_MissingField_DoesNotSend(t *testing.T) {
	rt, err := New([]Rule{{Field: "level", Pattern: ".*", Channel: "all"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch := rt.Channel("all")
	rt.Route(makeEntry(map[string]any{"msg": "no level field"}))

	select {
	case <-ch:
		t.Fatal("expected no entry when field is absent")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestFromConfig_EmptySlice_ReturnsNil(t *testing.T) {
	rt, err := FromConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt != nil {
		t.Fatal("expected nil router for empty config")
	}
}

func TestFromConfig_ValidRules_ReturnsRouter(t *testing.T) {
	rt, err := FromConfig([]RuleConfig{
		{Field: "level", Pattern: "warn|error", Channel: "alerts"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt == nil {
		t.Fatal("expected non-nil router")
	}
}
