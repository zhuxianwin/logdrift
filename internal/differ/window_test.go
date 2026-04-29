package differ_test

import (
	"testing"
	"time"

	"github.com/logdrift/internal/differ"
	"github.com/logdrift/internal/parser"
)

func entry(msg string) parser.Entry {
	return parser.Entry{Message: msg, Fields: map[string]interface{}{}}
}

func TestWindow_AddAndLatest(t *testing.T) {
	w := differ.NewWindow(5 * time.Second)
	w.Add("svcA", entry("first"))
	w.Add("svcA", entry("second"))

	got, ok := w.Latest("svcA")
	if !ok {
		t.Fatal("expected entry for svcA")
	}
	if got.Entry.Message != "second" {
		t.Errorf("expected 'second', got %q", got.Entry.Message)
	}
}

func TestWindow_MissingService(t *testing.T) {
	w := differ.NewWindow(5 * time.Second)
	_, ok := w.Latest("unknown")
	if ok {
		t.Error("expected no entry for unknown service")
	}
}

func TestWindow_Services(t *testing.T) {
	w := differ.NewWindow(5 * time.Second)
	w.Add("alpha", entry("a"))
	w.Add("beta", entry("b"))

	svcs := w.Services()
	if len(svcs) != 2 {
		t.Errorf("expected 2 services, got %d", len(svcs))
	}
}

func TestWindow_Eviction(t *testing.T) {
	w := differ.NewWindow(50 * time.Millisecond)
	w.Add("svcA", entry("old"))

	time.Sleep(80 * time.Millisecond)

	// Trigger eviction by adding a new entry to a different service.
	w.Add("svcB", entry("new"))

	// Now add to svcA to trigger eviction for its stale entries.
	w.Add("svcA", entry("fresh"))

	got, ok := w.Latest("svcA")
	if !ok {
		t.Fatal("expected fresh entry")
	}
	if got.Entry.Message != "fresh" {
		t.Errorf("expected 'fresh', got %q", got.Entry.Message)
	}
}

func TestWindow_DiffLatest(t *testing.T) {
	w := differ.NewWindow(5 * time.Second)
	w.Add("svcA", makeEntry(map[string]interface{}{"code": 200}))
	w.Add("svcB", makeEntry(map[string]interface{}{"code": 500}))

	ea, _ := w.Latest("svcA")
	eb, _ := w.Latest("svcB")

	result := differ.Diff(ea.Service, eb.Service, ea.Entry, eb.Entry)
	if result.Identical {
		t.Error("expected differences between svcA and svcB")
	}
}
