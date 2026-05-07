package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
	"github.com/yourorg/logdrift/internal/snapshot"
)

func makeEntry(msg string) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Message:   msg,
		Fields:    map[string]any{"service": "svc-a"},
		Raw:       []byte(`{"msg":"` + msg + `"}`),
	}
}

func TestNew_CreatesEmptySnapshot(t *testing.T) {
	s := snapshot.New("test-label")
	if s.Label != "test-label" {
		t.Fatalf("expected label 'test-label', got %q", s.Label)
	}
	if len(s.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(s.Entries))
	}
}

func TestAdd_AppendsEntries(t *testing.T) {
	s := snapshot.New("add-test")
	s.Add("svc-a", makeEntry("hello"))
	s.Add("svc-a", makeEntry("world"))
	s.Add("svc-b", makeEntry("other"))

	if len(s.Entries["svc-a"]) != 2 {
		t.Fatalf("expected 2 entries for svc-a, got %d", len(s.Entries["svc-a"]))
	}
	if len(s.Entries["svc-b"]) != 1 {
		t.Fatalf("expected 1 entry for svc-b, got %d", len(s.Entries["svc-b"]))
	}
}

func TestServices_ReturnsServiceNames(t *testing.T) {
	s := snapshot.New("svc-test")
	s.Add("alpha", makeEntry("a"))
	s.Add("beta", makeEntry("b"))

	svcs := s.Services()
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	orig := snapshot.New("roundtrip")
	orig.Add("svc-a", makeEntry("ping"))

	if err := orig.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Label != "roundtrip" {
		t.Errorf("expected label 'roundtrip', got %q", loaded.Label)
	}
	if len(loaded.Entries["svc-a"]) != 1 {
		t.Errorf("expected 1 entry for svc-a, got %d", len(loaded.Entries["svc-a"]))
	}
}

func TestLoad_MissingFile_ReturnsError(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSave_InvalidPath_ReturnsError(t *testing.T) {
	s := snapshot.New("fail")
	err := s.Save("/nonexistent/dir/snap.json")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestNew_SetsCreatedAt(t *testing.T) {
	before := time.Now().UTC()
	s := snapshot.New("time-test")
	after := time.Now().UTC()

	if s.CreatedAt.Before(before) || s.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", s.CreatedAt, before, after)
	}
	_ = os.Getenv("") // suppress unused import lint
}
