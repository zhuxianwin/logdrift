package differ_test

import (
	"strings"
	"testing"

	"github.com/logdrift/internal/differ"
	"github.com/logdrift/internal/parser"
)

func makeEntry(fields map[string]interface{}) parser.Entry {
	return parser.Entry{
		Level:   "info",
		Message: "test",
		Fields:  fields,
	}
}

func TestDiff_IdenticalEntries(t *testing.T) {
	a := makeEntry(map[string]interface{}{"key": "value", "code": 200})
	b := makeEntry(map[string]interface{}{"key": "value", "code": 200})

	result := differ.Diff("svcA", "svcB", a, b)

	if !result.Identical {
		t.Errorf("expected identical, got diffs: %v", result.Diffs)
	}
}

func TestDiff_ValueChanged(t *testing.T) {
	a := makeEntry(map[string]interface{}{"status": "ok"})
	b := makeEntry(map[string]interface{}{"status": "error"})

	result := differ.Diff("svcA", "svcB", a, b)

	if result.Identical {
		t.Fatal("expected differences")
	}
	if len(result.Diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result.Diffs))
	}
	if result.Diffs[0].Field != "status" {
		t.Errorf("expected field 'status', got %q", result.Diffs[0].Field)
	}
}

func TestDiff_MissingFieldInB(t *testing.T) {
	a := makeEntry(map[string]interface{}{"trace_id": "abc123"})
	b := makeEntry(map[string]interface{}{})

	result := differ.Diff("svcA", "svcB", a, b)

	if result.Identical {
		t.Fatal("expected differences")
	}
	if result.Diffs[0].Missing != "B" {
		t.Errorf("expected Missing=B, got %q", result.Diffs[0].Missing)
	}
}

func TestDiff_MissingFieldInA(t *testing.T) {
	a := makeEntry(map[string]interface{}{})
	b := makeEntry(map[string]interface{}{"region": "us-east-1"})

	result := differ.Diff("svcA", "svcB", a, b)

	if result.Diffs[0].Missing != "A" {
		t.Errorf("expected Missing=A, got %q", result.Diffs[0].Missing)
	}
}

func TestResult_Format_Identical(t *testing.T) {
	r := differ.Result{ServiceA: "a", ServiceB: "b", Identical: true}
	out := r.Format()
	if !strings.Contains(out, "identical") {
		t.Errorf("expected 'identical' in output, got: %q", out)
	}
}

func TestResult_Format_WithDiffs(t *testing.T) {
	r := differ.Result{
		ServiceA: "a",
		ServiceB: "b",
		Diffs: []differ.FieldDiff{
			{Field: "level", ValueA: "info", ValueB: "warn"},
		},
	}
	out := r.Format()
	if !strings.Contains(out, "level") {
		t.Errorf("expected 'level' in output, got: %q", out)
	}
	if !strings.Contains(out, "→") {
		t.Errorf("expected arrow in output, got: %q", out)
	}
}
