package pipeline

import (
	"context"

	"github.com/user/logdrift/internal/differ"
	"github.com/user/logdrift/internal/formatter"
	"github.com/user/logdrift/internal/parser"
	"github.com/user/logdrift/internal/tailer"
)

// Config holds the configuration for a pipeline run.
type Config struct {
	// Services maps service names to their log file paths.
	Services map[string]string
	// WindowSize is the number of recent entries to keep per service.
	WindowSize int
	// Plain disables colour output when true.
	Plain bool
	// Output is the writer that receives formatted diff lines.
	Output interface {
		WriteString(s string) (int, error)
	}
}

// Run wires together the tailer manager, parser, differ, and formatter,
// streaming diff output until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	windowSize := cfg.WindowSize
	if windowSize <= 0 {
		windowSize = 50
	}

	serviceNames := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		serviceNames = append(serviceNames, name)
	}

	win := differ.NewWindow(windowSize)
	mgr := tailer.NewManager(cfg.Services)

	lines, err := mgr.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tagged, ok := <-lines:
			if !ok {
				return nil
			}
			entry, parseErr := parser.Parse(tagged.Service, tagged.Line)
			if parseErr != nil {
				continue
			}
			win.Add(tagged.Service, entry)

			latest := win.Latest(serviceNames)
			if len(latest) < 2 {
				continue
			}

			deltas := differ.Diff(latest)
			if len(deltas) == 0 {
				continue
			}

			var output string
			if cfg.Plain {
				output = formatter.FormatPlain(serviceNames, deltas)
			} else {
				output = formatter.Format(serviceNames, deltas)
			}
			if output != "" {
				cfg.Output.WriteString(output + "\n")
			}
		}
	}
}
