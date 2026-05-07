package redact_test

import (
	"testing"

	"github.com/yourorg/logdrift/internal/parser"
	"github.com/yourorg/logdrift/internal/redact"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Service: "svc",
		Message: "test",
		Fields:  fields,
	}
}

func TestNew_EmptyFieldName_ReturnsError(t *testing.T) {
	_, err := redact.New([]redact.Rule{{Field: ""}})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := redact.New([]redact.Rule{{Field: "token", Pattern: "[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestApply_NoRules_ReturnsEntryUnchanged(t *testing.T) {
	r, _ := redact.New(nil)
	e := makeEntry(map[string]any{"password": "secret"})
	out := r.Apply(e)
	if out.Fields["password"] != "secret" {
		t.Errorf("expected unchanged value, got %v", out.Fields["password"])
	}
}

func TestApply_FullFieldRedaction_MasksValue(t *testing.T) {
	r, err := redact.New([]redact.Rule{{Field: "password"}})
	if err != nil {
		t.Fatal(err)
	}
	e := makeEntry(map[string]any{"password": "hunter2"})
	out := r.Apply(e)
	if out.Fields["password"] == "hunter2" {
		t.Error("expected password to be masked")
	}
}

func TestApply_PatternRedaction_ReplacesMatch(t *testing.T) {
	r, err := redact.New([]redact.Rule{{Field: "msg", Pattern: `\d{4}-\d{4}-\d{4}-\d{4}`}})
	if err != nil {
		t.Fatal(err)
	}
	e := makeEntry(map[string]any{"msg": "card: 1234-5678-9012-3456 processed"})
	out := r.Apply(e)
	v, _ := out.Fields["msg"].(string)
	if v == "card: 1234-5678-9012-3456 processed" {
		t.Error("expected card number to be redacted")
	}
	if v == "" {
		t.Error("expected surrounding text to be preserved")
	}
}

func TestApply_MissingField_NoChange(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "secret"}})
	e := makeEntry(map[string]any{"other": "value"})
	out := r.Apply(e)
	if _, ok := out.Fields["secret"]; ok {
		t.Error("did not expect secret field to appear")
	}
}

func TestApply_NonStringField_RedactsWithPlaceholder(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "pin"}})
	e := makeEntry(map[string]any{"pin": 1234})
	out := r.Apply(e)
	if out.Fields["pin"] == 1234 {
		t.Error("expected non-string field to be redacted")
	}
}

func TestApply_OriginalEntryUnmodified(t *testing.T) {
	r, _ := redact.New([]redact.Rule{{Field: "token"}})
	e := makeEntry(map[string]any{"token": "abc123"})
	_ = r.Apply(e)
	if e.Fields["token"] != "abc123" {
		t.Error("original entry should not be modified")
	}
}
