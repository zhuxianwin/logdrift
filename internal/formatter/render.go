package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/logdrift/internal/differ"
)

// Renderer writes formatted diff output to a writer.
type Renderer struct {
	w      io.Writer
	colors ColorScheme
	stats  *SummaryStats
}

// NewRenderer creates a Renderer that writes to w using the given color scheme.
func NewRenderer(w io.Writer, colors ColorScheme) *Renderer {
	return &Renderer{
		w:      w,
		colors: colors,
		stats:  NewSummaryStats(),
	}
}

// Write formats a batch of deltas and writes it to the underlying writer.
// It records statistics and returns any write error.
func (r *Renderer) Write(services []string, deltas []differ.Delta) error {
	r.stats.Record(deltas)

	lines := FormatPlain(services, deltas, r.colors)
	if len(lines) == 0 {
		return nil
	}

	_, err := fmt.Fprintln(r.w, strings.Join(lines, "\n"))
	return err
}

// Summary writes the session summary to the underlying writer.
func (r *Renderer) Summary() error {
	_, err := fmt.Fprintln(r.w, r.stats.Render())
	return err
}

// Stats returns the accumulated SummaryStats for the session.
func (r *Renderer) Stats() *SummaryStats {
	return r.stats
}
