package fieldmask_test

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/fieldmask"
	"github.com/user/logdrift/internal/parser"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Fields:    fields,
	}
}

func TestNew_EmptyFields_ReturnsError(t *testing.T) {
	_, err := fieldmask.New([]string{})
	if err == nil {
		t.Fatal("expected error for empty fields slice")
	}
}

func TestNew_BlankFieldName_ReturnsError(t *testing.T) {
	_, err := fieldmask.New([]string{"valid", "  "})
	if err == nil {
		t.Fatal("expected error for blank field name")
	}
}

func TestNew_ValidFields_ReturnsNoError(t *testing.T) {
	_, err := fieldmask.New([]string{"level", "request_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_KeepsAllowedFields(t *testing.T) {
	m, _ := fieldmask.New([]string{"request_id", "status"})
	e := makeEntry(map[string]any{
		"request_id": "abc-123",
		"status":     200,
		"latency_ms": 42,
	})
	out := m.Apply(e)
	if _, ok := out.Fields["request_id"]; !ok {
		t.Error("expected request_id to be present")
	}
	if _, ok := out.Fields["status"]; !ok {
		t.Error("expected status to be present")
	}
	if _, ok := out.Fields["latency_ms"]; ok {
		t.Error("expected latency_ms to be dropped")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	m, _ := fieldmask.New([]string{"status"})
	e := makeEntry(map[string]any{"status": 200, "extra": "keep"})
	_ = m.Apply(e)
	if _, ok := e.Fields["extra"]; !ok {
		t.Error("original entry should not be mutated")
	}
}

func TestApply_PreservesEntryMetadata(t *testing.T) {
	m, _ := fieldmask.New([]string{"x"})
	e := makeEntry(map[string]any{"x": 1})
	e.Service = "payments"
	e.Level = "warn"
	out := m.Apply(e)
	if out.Service != "payments" {
		t.Errorf("expected service payments, got %s", out.Service)
	}
	if out.Level != "warn" {
		t.Errorf("expected level warn, got %s", out.Level)
	}
}

func TestContains_KnownField_ReturnsTrue(t *testing.T) {
	m, _ := fieldmask.New([]string{"user_id"})
	if !m.Contains("user_id") {
		t.Error("expected Contains to return true for user_id")
	}
}

func TestContains_UnknownField_ReturnsFalse(t *testing.T) {
	m, _ := fieldmask.New([]string{"user_id"})
	if m.Contains("trace_id") {
		t.Error("expected Contains to return false for trace_id")
	}
}

func TestFields_ReturnsAllowedNames(t *testing.T) {
	m, _ := fieldmask.New([]string{"a", "b", "c"})
	fields := m.Fields()
	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}
}
