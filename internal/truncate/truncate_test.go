package truncate_test

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
	"github.com/yourorg/logdrift/internal/truncate"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Fields:    fields,
	}
}

func TestNew_EmptyRules_ReturnsError(t *testing.T) {
	_, err := truncate.New(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := truncate.New([]truncate.Rule{{Field: "", MaxLen: 10}})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNew_ZeroMaxLen_ReturnsError(t *testing.T) {
	_, err := truncate.New([]truncate.Rule{{Field: "msg", MaxLen: 0}})
	if err == nil {
		t.Fatal("expected error for zero max_len")
	}
}

func TestNew_NegativeMaxLen_ReturnsError(t *testing.T) {
	_, err := truncate.New([]truncate.Rule{{Field: "msg", MaxLen: -5}})
	if err == nil {
		t.Fatal("expected error for negative max_len")
	}
}

func TestApply_ShortValue_Unchanged(t *testing.T) {
	tr, _ := truncate.New([]truncate.Rule{{Field: "msg", MaxLen: 20}})
	entry := makeEntry(map[string]any{"msg": "hello"})
	out := tr.Apply(entry)
	if out.Fields["msg"] != "hello" {
		t.Fatalf("expected 'hello', got %q", out.Fields["msg"])
	}
}

func TestApply_LongValue_Truncated(t *testing.T) {
	tr, _ := truncate.New([]truncate.Rule{{Field: "msg", MaxLen: 5}})
	entry := makeEntry(map[string]any{"msg": "hello world"})
	out := tr.Apply(entry)
	if out.Fields["msg"] != "hello" {
		t.Fatalf("expected 'hello', got %q", out.Fields["msg"])
	}
}

func TestApply_MissingField_NoChange(t *testing.T) {
	tr, _ := truncate.New([]truncate.Rule{{Field: "missing", MaxLen: 5}})
	entry := makeEntry(map[string]any{"msg": "hello world"})
	out := tr.Apply(entry)
	if out.Fields["msg"] != "hello world" {
		t.Fatalf("unexpected change to unrelated field: %q", out.Fields["msg"])
	}
}

func TestApply_NonStringField_Unchanged(t *testing.T) {
	tr, _ := truncate.New([]truncate.Rule{{Field: "count", MaxLen: 2}})
	entry := makeEntry(map[string]any{"count": 42})
	out := tr.Apply(entry)
	if out.Fields["count"] != 42 {
		t.Fatalf("expected 42, got %v", out.Fields["count"])
	}
}

func TestApply_OriginalEntryUnmodified(t *testing.T) {
	tr, _ := truncate.New([]truncate.Rule{{Field: "msg", MaxLen: 3}})
	entry := makeEntry(map[string]any{"msg": "hello"})
	_ = tr.Apply(entry)
	if entry.Fields["msg"] != "hello" {
		t.Fatal("original entry was mutated")
	}
}
