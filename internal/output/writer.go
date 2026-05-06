package output

import (
	"io"
	"os"
	"sync"
)

// Writer wraps an io.Writer with thread-safe writes and optional line buffering.
type Writer struct {
	mu  sync.Mutex
	out io.Writer
}

// New returns a Writer backed by the given io.Writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Writer {
	if w == nil {
		w = os.Stdout
	}
	return &Writer{out: w}
}

// Write writes p to the underlying writer under a mutex.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.out.Write(p)
}

// WriteLine writes s followed by a newline to the underlying writer.
func (w *Writer) WriteLine(s string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := io.WriteString(w.out, s+"\n")
	return err
}

// WriteLines writes each string in lines as a separate line.
func (w *Writer) WriteLines(lines []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, line := range lines {
		if _, err := io.WriteString(w.out, line+"\n"); err != nil {
			return err
		}
	}
	return nil
}

// Unwrap returns the underlying io.Writer.
func (w *Writer) Unwrap() io.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.out
}
