package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/yourorg/logdrift/internal/differ"
)

// DiffLine represents a single rendered output line for a diff result.
type DiffLine struct {
	Service string
	Text    string
}

var (
	addedColor   = color.New(color.FgGreen)
	removedColor = color.New(color.FgRed)
	changedColor = color.New(color.FgYellow)
	labelColor   = color.New(color.FgCyan, color.Bold)
)

// Format converts a slice of differ.Delta values into human-readable lines.
func Format(serviceA, serviceB string, ts time.Time, deltas []differ.Delta) []string {
	if len(deltas) == 0 {
		return nil
	}

	lines := make([]string, 0, len(deltas)+1)

	header := labelColor.Sprintf(
		"[%s] diff %s ↔ %s",
		ts.Format(time.RFC3339),
		serviceA,
		serviceB,
	)
	lines = append(lines, header)

	for _, d := range deltas {
		var line string
		switch d.Type {
		case differ.DeltaAdded:
			line = addedColor.Sprintf("  + %-20s %v", d.Field, d.ValueB)
		case differ.DeltaRemoved:
			line = removedColor.Sprintf("  - %-20s %v", d.Field, d.ValueA)
		case differ.DeltaChanged:
			line = changedColor.Sprintf("  ~ %-20s %v → %v", d.Field, d.ValueA, d.ValueB)
		}
		lines = append(lines, line)
	}

	return lines
}

// FormatPlain returns plain-text (no ANSI colour) diff output suitable for
// piping or log files.
func FormatPlain(serviceA, serviceB string, ts time.Time, deltas []differ.Delta) []string {
	if len(deltas) == 0 {
		return nil
	}

	lines := make([]string, 0, len(deltas)+1)
	lines = append(lines, fmt.Sprintf(
		"[%s] diff %s <-> %s",
		ts.Format(time.RFC3339), serviceA, serviceB,
	))

	for _, d := range deltas {
		var prefix string
		switch d.Type {
		case differ.DeltaAdded:
			prefix = "+"
		case differ.DeltaRemoved:
			prefix = "-"
		case differ.DeltaChanged:
			prefix = "~"
		}
		lines = append(lines, fmt.Sprintf("  %s %-20s %v",
			prefix, d.Field, formatValues(d)))
	}

	return lines
}

func formatValues(d differ.Delta) string {
	switch d.Type {
	case differ.DeltaAdded:
		return fmt.Sprintf("%v", d.ValueB)
	case differ.DeltaRemoved:
		return fmt.Sprintf("%v", d.ValueA)
	case differ.DeltaChanged:
		return fmt.Sprintf("%v -> %v", d.ValueA, d.ValueB)
	}
	return ""
}

// Join is a convenience helper that concatenates formatted lines with newlines.
func Join(lines []string) string {
	return strings.Join(lines, "\n")
}
