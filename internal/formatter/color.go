package formatter

import "fmt"

// ANSI color codes used for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// ColorScheme controls whether ANSI escape codes are emitted.
type ColorScheme struct {
	Enabled bool
}

// DefaultColorScheme returns a ColorScheme with color enabled.
func DefaultColorScheme() ColorScheme {
	return ColorScheme{Enabled: true}
}

// NoColorScheme returns a ColorScheme with color disabled.
func NoColorScheme() ColorScheme {
	return ColorScheme{Enabled: false}
}

// Changed wraps s in red to indicate a changed value.
func (c ColorScheme) Changed(s string) string {
	if !c.Enabled {
		return s
	}
	return fmt.Sprintf("%s%s%s", colorRed, s, colorReset)
}

// Added wraps s in green to indicate an added value.
func (c ColorScheme) Added(s string) string {
	if !c.Enabled {
		return s
	}
	return fmt.Sprintf("%s%s%s", colorGreen, s, colorReset)
}

// Missing wraps s in yellow to indicate a missing value.
func (c ColorScheme) Missing(s string) string {
	if !c.Enabled {
		return s
	}
	return fmt.Sprintf("%s%s%s", colorYellow, s, colorReset)
}

// Header wraps s in bold cyan for section headers.
func (c ColorScheme) Header(s string) string {
	if !c.Enabled {
		return s
	}
	return fmt.Sprintf("%s%s%s%s", colorBold, colorCyan, s, colorReset)
}
