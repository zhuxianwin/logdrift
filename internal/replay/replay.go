package replay

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

// Speed controls how fast the replay runs relative to original timestamps.
// 1.0 = real time, 2.0 = double speed, 0 = no delay (dump all at once).
type Options struct {
	Speed   float64
	NoDelay bool
}

// Line represents a replayed log line with its parsed entry.
type Line struct {
	Service string
	Raw     string
	Entry   parser.Entry
}

// Run reads lines from r, parses timestamps, and emits them on the returned
// channel respecting the relative timing between entries. The channel is
// closed when r is exhausted or ctx is cancelled.
func Run(ctx context.Context, service string, r io.Reader, opts Options) (<-chan Line, error) {
	ch := make(chan Line, 64)

	go func() {
		defer close(ch)

		scanner := bufio.NewScanner(r)
		var prevLogTime time.Time
		var prevWall time.Time

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			raw := scanner.Text()
			entry, err := parser.Parse(raw)
			if err != nil {
				continue
			}

			if !opts.NoDelay && opts.Speed > 0 && !entry.Time.IsZero() {
				if !prevLogTime.IsZero() {
					logGap := entry.Time.Sub(prevLogTime)
					wallGap := time.Duration(float64(logGap) / opts.Speed)
					deadline := prevWall.Add(wallGap)
					waitFor := time.Until(deadline)
					if waitFor > 0 {
						select {
						case <-time.After(waitFor):
						case <-ctx.Done():
							return
						}
					}
				}
				prevLogTime = entry.Time
				prevWall = time.Now()
			}

			ch <- Line{Service: service, Raw: raw, Entry: entry}
		}
	}()

	return ch, nil
}

// OpenFile is a convenience wrapper that opens a file and calls Run.
func OpenFile(ctx context.Context, service, path string, opts Options) (<-chan Line, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	ch, err := Run(ctx, service, f, opts)
	if err != nil {
		f.Close()
		return nil, err
	}
	// file will be GC'd when reader is exhausted; acceptable for CLI tool.
	return ch, nil
}
