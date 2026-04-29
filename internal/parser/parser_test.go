package parser

import (
	"testing"
	"time"
)

func TestParse_ValidJSON(t *testing.T) {
	line := `{"level":"info","msg":"server started","time":"2024-01-15T10:00:00Z"}`
	entry, err := Parse("api", line)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Service != "api" {
		t.Errorf("expected service 'api', got %q", entry.Service)
	}
	if entry.Level != "info" {
		t.Errorf("expected level 'info', got %q", entry.Level)
	}
	if entry.Message != "server started" {
		t.Errorf("expected message 'server started', got %q", entry.Message)
	}
	expected := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, entry.Timestamp)
	}
}

func TestParse_NonJSON(t *testing.T) {
	line := "plain text log line"
	entry, err := Parse("worker", line)
	if err == nil {
		t.Fatal("expected error for non-JSON line")
	}
	if entry.Raw != line {
		t.Errorf("expected raw to be preserved, got %q", entry.Raw)
	}
	if entry.Service != "worker" {
		t.Errorf("expected service 'worker', got %q", entry.Service)
	}
}

func TestParse_AlternativeFieldNames(t *testing.T) {
	line := `{"severity":"error","message":"disk full","timestamp":"2024-03-01T08:30:00Z"}`
	entry, err := Parse("storage", line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "error" {
		t.Errorf("expected level 'error', got %q", entry.Level)
	}
	if entry.Message != "disk full" {
		t.Errorf("expected message 'disk full', got %q", entry.Message)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParse_UnixTimestamp(t *testing.T) {
	line := `{"level":"debug","msg":"tick","ts":1700000000}`
	entry, err := Parse("scheduler", line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp from unix epoch")
	}
	if entry.Timestamp.Unix() != 1700000000 {
		t.Errorf("expected unix 1700000000, got %d", entry.Timestamp.Unix())
	}
}

func TestParse_MissingOptionalFields(t *testing.T) {
	line := `{"custom_key":"value"}`
	entry, err := Parse("svc", line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "" {
		t.Errorf("expected empty level, got %q", entry.Level)
	}
	if entry.Message != "" {
		t.Errorf("expected empty message, got %q", entry.Message)
	}
	if !entry.Timestamp.IsZero() {
		t.Errorf("expected zero timestamp, got %v", entry.Timestamp)
	}
}
