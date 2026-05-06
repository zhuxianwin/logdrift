package output

import (
	"fmt"
	"io"
	"os"

	"github.com/user/logdrift/internal/formatter"
)

// Format enumerates supported output formats.
type Format string

const (
	FormatColor Format = "color"
	FormatPlain Format = "plain"
)

// Options holds configuration for building an output Writer and color scheme.
type Options struct {
	Format  Format
	OutFile string // if non-empty, write to this path instead of stdout
}

// Build constructs a Writer and ColorScheme from Options.
// The caller is responsible for closing the returned io.Closer if non-nil.
func Build(opts Options) (*Writer, formatter.ColorScheme, io.Closer, error) {
	var (
		w      io.Writer = os.Stdout
		closer io.Closer
	)

	if opts.OutFile != "" {
		f, err := os.OpenFile(opts.OutFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, formatter.ColorScheme{}, nil, fmt.Errorf("output: open file %q: %w", opts.OutFile, err)
		}
		w = f
		closer = f
	}

	var scheme formatter.ColorScheme
	switch opts.Format {
	case FormatColor, "":
		scheme = formatter.DefaultColorScheme()
	case FormatPlain:
		scheme = formatter.NoColorScheme()
	default:
		return nil, formatter.ColorScheme{}, closer, fmt.Errorf("output: unknown format %q", opts.Format)
	}

	return New(w), scheme, closer, nil
}
