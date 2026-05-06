package formatter

import (
	"strings"
	"testing"
)

func TestColorScheme_Disabled_ReturnsPlainText(t *testing.T) {
	c := NoColorScheme()
	for _, fn := range []func(string) string{
		c.Changed, c.Added, c.Missing, c.Header,
	} {
		if got := fn("hello"); got != "hello" {
			t.Errorf("expected plain \"hello\", got %q", got)
		}
	}
}

func TestColorScheme_Enabled_WrapsWithEscapeCodes(t *testing.T) {
	c := DefaultColorScheme()
	tests := []struct {
		name string
		fn   func(string) string
		code string
	}{
		{"Changed", c.Changed, colorRed},
		{"Added", c.Added, colorGreen},
		{"Missing", c.Missing, colorYellow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("val")
			if !strings.HasPrefix(got, tt.code) {
				t.Errorf("%s: expected prefix %q, got %q", tt.name, tt.code, got)
			}
			if !strings.HasSuffix(got, colorReset) {
				t.Errorf("%s: expected suffix reset code, got %q", tt.name, got)
			}
			if !strings.Contains(got, "val") {
				t.Errorf("%s: expected original value in output, got %q", tt.name, got)
			}
		})
	}
}

func TestColorScheme_Header_UsesBoldCyan(t *testing.T) {
	c := DefaultColorScheme()
	got := c.Header("Services")
	if !strings.Contains(got, colorBold) {
		t.Errorf("Header: expected bold code, got %q", got)
	}
	if !strings.Contains(got, colorCyan) {
		t.Errorf("Header: expected cyan code, got %q", got)
	}
	if !strings.Contains(got, "Services") {
		t.Errorf("Header: expected original text, got %q", got)
	}
}

func TestColorScheme_Enabled_EmptyString(t *testing.T) {
	c := DefaultColorScheme()
	got := c.Changed("")
	if !strings.HasPrefix(got, colorRed) {
		t.Errorf("expected color prefix for empty string, got %q", got)
	}
}
