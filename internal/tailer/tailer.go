package tailer

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

// Line represents a single log line with its source label.
type Line struct {
	Source  string
	Content string
	Ts      time.Time
}

// Tailer tails a file and emits lines to a channel.
type Tailer struct {
	Source string
	Path   string
	Out    chan<- Line
}

// New creates a new Tailer for the given source label and file path.
func New(source, path string, out chan<- Line) *Tailer {
	return &Tailer{
		Source: source,
		Path:   path,
		Out:    out,
	}
}

// Tail opens the file, seeks to the end, and streams new lines until ctx is cancelled.
func (t *Tailer) Tail(ctx context.Context) error {
	f, err := os.Open(t.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}

		if len(line) > 0 {
			t.Out <- Line{
				Source:  t.Source,
				Content: line,
				Ts:      time.Now(),
			}
		}
	}
}
