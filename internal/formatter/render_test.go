package formatter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/differ"
)

func TestRenderer_WriteNoDeltas_WritesNothing(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, NoColorScheme())

	if err := r.Write([]string{"svc-a", "svc-b"}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestRenderer_WriteDeltas_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, NoColorScheme())

	ds := []differ.Delta{
		{Field: "level", A: "info", B: "error", Changed: true},
	}
	services := []string{"alpha", "beta"}

	if err := r.Write(services, ds); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, svc := range services {
		if !strings.Contains(out, svc) {
			t.Errorf("expected output to contain service %q, got:\n%s", svc, out)
		}
	}
}

func TestRenderer_WriteDeltas_RecordsStats(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, NoColorScheme())

	ds := []differ.Delta{
		{Field: "msg", A: "hello", B: "world", Changed: true},
	}

	_ = r.Write([]string{"x", "y"}, ds)

	stats := r.Stats()
	if stats.TotalBatches != 1 {
		t.Errorf("expected 1 batch recorded, got %d", stats.TotalBatches)
	}
	if stats.DriftBatches != 1 {
		t.Errorf("expected 1 drift batch, got %d", stats.DriftBatches)
	}
}

func TestRenderer_Summary_ContainsDriftWord(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, NoColorScheme())

	_ = r.Write([]string{"a", "b"}, []differ.Delta{
		{Field: "level", A: "debug", B: "warn", Changed: true},
	})
	buf.Reset()

	if err := r.Summary(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(strings.ToLower(out), "drift") {
		t.Errorf("expected summary to mention drift, got: %s", out)
	}
}
