package enricher

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/parser"
)

func makeEntry(fields map[string]interface{}) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Service:   "svc",
		Fields:    fields,
	}
}

func TestNew_EmptySourceField_ReturnsError(t *testing.T) {
	_, err := New([]Rule{{SourceField: "", DestField: "dst"}})
	if err == nil {
		t.Fatal("expected error for empty source_field")
	}
}

func TestNew_EmptyDestField_ReturnsError(t *testing.T) {
	_, err := New([]Rule{{SourceField: "src", DestField: ""}})
	if err == nil {
		t.Fatal("expected error for empty dest_field")
	}
}

func TestNew_ValidRules_ReturnsEnricher(t *testing.T) {
	e, err := New([]Rule{{SourceField: "env", DestField: "env_tag"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil enricher")
	}
}

func TestApply_NoRules_ReturnsEntryUnchanged(t *testing.T) {
	e, _ := New([]Rule{})
	entry := makeEntry(map[string]interface{}{"key": "val"})
	out := e.Apply(entry)
	if len(out.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(out.Fields))
	}
}

func TestApply_CopiesSourceToDestField(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "env", DestField: "env_tag"}})
	entry := makeEntry(map[string]interface{}{"env": "production"})
	out := e.Apply(entry)
	if out.Fields["env_tag"] != "production" {
		t.Fatalf("expected 'production', got %v", out.Fields["env_tag"])
	}
}

func TestApply_PrefixPrepended(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "region", DestField: "region_tag", Prefix: "region:"}})
	entry := makeEntry(map[string]interface{}{"region": "us-east-1"})
	out := e.Apply(entry)
	if out.Fields["region_tag"] != "region:us-east-1" {
		t.Fatalf("unexpected value: %v", out.Fields["region_tag"])
	}
}

func TestApply_UppercaseTransforms(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "level", DestField: "level_upper", Uppercase: true}})
	entry := makeEntry(map[string]interface{}{"level": "warn"})
	out := e.Apply(entry)
	if out.Fields["level_upper"] != "WARN" {
		t.Fatalf("expected 'WARN', got %v", out.Fields["level_upper"])
	}
}

func TestApply_DestFieldExists_NotOverwritten(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "env", DestField: "tag"}})
	entry := makeEntry(map[string]interface{}{"env": "prod", "tag": "existing"})
	out := e.Apply(entry)
	if out.Fields["tag"] != "existing" {
		t.Fatalf("expected existing value to be preserved, got %v", out.Fields["tag"])
	}
}

func TestApply_MissingSourceField_SkipsRule(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "missing", DestField: "dst"}})
	entry := makeEntry(map[string]interface{}{"other": "val"})
	out := e.Apply(entry)
	if _, ok := out.Fields["dst"]; ok {
		t.Fatal("expected dst field to be absent")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	e, _ := New([]Rule{{SourceField: "env", DestField: "env_tag"}})
	orig := map[string]interface{}{"env": "staging"}
	entry := makeEntry(orig)
	e.Apply(entry)
	if _, ok := orig["env_tag"]; ok {
		t.Fatal("original fields map was mutated")
	}
}
