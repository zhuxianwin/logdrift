package differ

import (
	"fmt"
	"sort"
	"strings"

	"github.com/logdrift/internal/parser"
)

// FieldDiff represents a difference in a single field between two log entries.
type FieldDiff struct {
	Field    string
	ValueA   interface{}
	ValueB   interface{}
	Missing  string // "A" or "B" if field is absent in one entry
}

// Result holds the comparison outcome between two parsed log entries.
type Result struct {
	ServiceA string
	ServiceB string
	Diffs    []FieldDiff
	Identical bool
}

// Diff compares two parsed log entries from different services and returns a Result.
func Diff(serviceA, serviceB string, entryA, entryB parser.Entry) Result {
	diffs := compareFields(entryA.Fields, entryB.Fields)
	return Result{
		ServiceA:  serviceA,
		ServiceB:  serviceB,
		Diffs:     diffs,
		Identical: len(diffs) == 0,
	}
}

func compareFields(a, b map[string]interface{}) []FieldDiff {
	var diffs []FieldDiff
	keys := unionKeys(a, b)

	for _, key := range keys {
		valA, okA := a[key]
		valB, okB := b[key]

		switch {
		case okA && !okB:
			diffs = append(diffs, FieldDiff{Field: key, ValueA: valA, Missing: "B"})
		case !okA && okB:
			diffs = append(diffs, FieldDiff{Field: key, ValueB: valB, Missing: "A"})
		case fmt.Sprintf("%v", valA) != fmt.Sprintf("%v", valB):
			diffs = append(diffs, FieldDiff{Field: key, ValueA: valA, ValueB: valB})
		}
	}
	return diffs
}

func unionKeys(a, b map[string]interface{}) []string {
	seen := make(map[string]struct{})
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Format returns a human-readable summary of the diff result.
func (r Result) Format() string {
	if r.Identical {
		return fmt.Sprintf("[=] %s vs %s: identical", r.ServiceA, r.ServiceB)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[~] %s vs %s:\n", r.ServiceA, r.ServiceB)
	for _, d := range r.Diffs {
		switch d.Missing {
		case "A":
			fmt.Fprintf(&sb, "  + %s: %v (only in %s)\n", d.Field, d.ValueB, r.ServiceB)
		case "B":
			fmt.Fprintf(&sb, "  - %s: %v (only in %s)\n", d.Field, d.ValueA, r.ServiceA)
		default:
			fmt.Fprintf(&sb, "  ~ %s: %v → %v\n", d.Field, d.ValueA, d.ValueB)
		}
	}
	return sb.String()
}
