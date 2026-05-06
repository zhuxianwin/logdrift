package output_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/user/logdrift/internal/output"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	w := output.New(nil)
	if w == nil {
		t.Fatal("expected non-nil Writer")
	}
}

func TestWriteLine_AppendsNewline(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf)
	if err := w.WriteLine("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Errorf("expected %q, got %q", "hello\n", got)
	}
}

func TestWriteLines_WritesEachLine(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf)
	lines := []string{"alpha", "beta", "gamma"}
	if err := w.WriteLines(lines); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	for _, l := range lines {
		if !strings.Contains(got, l+"\n") {
			t.Errorf("expected output to contain %q", l+"\n")
		}
	}
}

func TestWrite_RawBytes(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf)
	p := []byte("raw bytes")
	n, err := w.Write(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(p) {
		t.Errorf("expected %d bytes written, got %d", len(p), n)
	}
	if buf.String() != "raw bytes" {
		t.Errorf("unexpected content: %q", buf.String())
	}
}

func TestUnwrap_ReturnsUnderlying(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf)
	if w.Unwrap() != &buf {
		t.Error("Unwrap did not return the original writer")
	}
}

func TestWriteLine_ConcurrentSafe(t *testing.T) {
	var buf syncBuffer
	w := output.New(&buf)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = w.WriteLine("concurrent")
		}()
	}
	wg.Wait()
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}
