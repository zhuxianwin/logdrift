// Package fieldmask provides a way to restrict diff output to a specific
// set of JSON fields, ignoring all others during comparison.
package fieldmask

import (
	"errors"
	"strings"

	"github.com/user/logdrift/internal/parser"
)

// Mask holds the set of allowed top-level field names.
type Mask struct {
	allowed map[string]struct{}
}

// New creates a Mask from a slice of field names.
// Field names are trimmed of whitespace. Duplicates are silently ignored.
// Returns an error if fields is empty or any entry is blank after trimming.
func New(fields []string) (*Mask, error) {
	if len(fields) == 0 {
		return nil, errors.New("fieldmask: at least one field name is required")
	}
	allowed := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			return nil, errors.New("fieldmask: field name must not be blank")
		}
		allowed[f] = struct{}{}
	}
	return &Mask{allowed: allowed}, nil
}

// Apply returns a copy of entry whose Fields map contains only the keys
// present in the mask. Fields not in the mask are dropped. The original
// entry is never mutated.
func (m *Mask) Apply(e parser.Entry) parser.Entry {
	out := parser.Entry{
		Service:   e.Service,
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Message:   e.Message,
		Raw:       e.Raw,
		Fields:    make(map[string]any, len(m.allowed)),
	}
	for k, v := range e.Fields {
		if _, ok := m.allowed[k]; ok {
			out.Fields[k] = v
		}
	}
	return out
}

// Contains reports whether the given field name is in the mask.
func (m *Mask) Contains(field string) bool {
	_, ok := m.allowed[field]
	return ok
}

// Fields returns the allowed field names as a sorted slice.
func (m *Mask) Fields() []string {
	out := make([]string, 0, len(m.allowed))
	for k := range m.allowed {
		out = append(out, k)
	}
	return out
}
