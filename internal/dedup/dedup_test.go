package dedup

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

func makeEntry(service string, fields map[string]any) parser.Entry {
	return parser.Entry{
		Service: service,
		Fields:  fields,
	}
}

func TestNew_NegativeWindow_ReturnsError(t *testing.T) {
	_, err := New(-time.Second)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestIsDuplicate_FirstSeen_ReturnsFalse(t *testing.T) {
	d, _ := New(time.Minute)
	e := makeEntry("svc-a", map[string]any{"msg": "hello"})
	if d.IsDuplicate(e) {
		t.Error("first occurrence should not be a duplicate")
	}
}

func TestIsDuplicate_SecondSeen_ReturnsTrue(t *testing.T) {
	d, _ := New(time.Minute)
	e := makeEntry("svc-a", map[string]any{"msg": "hello"})
	d.IsDuplicate(e)
	if !d.IsDuplicate(e) {
		t.Error("second occurrence should be a duplicate")
	}
}

func TestIsDuplicate_DifferentService_ReturnsFalse(t *testing.T) {
	d, _ := New(time.Minute)
	a := makeEntry("svc-a", map[string]any{"msg": "hello"})
	b := makeEntry("svc-b", map[string]any{"msg": "hello"})
	d.IsDuplicate(a)
	if d.IsDuplicate(b) {
		t.Error("different service should not be a duplicate")
	}
}

func TestIsDuplicate_DifferentFields_ReturnsFalse(t *testing.T) {
	d, _ := New(time.Minute)
	a := makeEntry("svc-a", map[string]any{"msg": "hello"})
	b := makeEntry("svc-a", map[string]any{"msg": "world"})
	d.IsDuplicate(a)
	if d.IsDuplicate(b) {
		t.Error("different fields should not be a duplicate")
	}
}

func TestIsDuplicate_AfterWindowExpiry_ReturnsFalse(t *testing.T) {
	d, _ := New(time.Second)
	now := time.Now()
	d.nowFunc = func() time.Time { return now }

	e := makeEntry("svc-a", map[string]any{"msg": "hello"})
	d.IsDuplicate(e)

	// Advance time beyond the window.
	d.nowFunc = func() time.Time { return now.Add(2 * time.Second) }
	if d.IsDuplicate(e) {
		t.Error("entry should not be a duplicate after window expiry")
	}
}
