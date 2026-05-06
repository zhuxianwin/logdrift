package formatter

import (
	"fmt"
	"strings"

	"github.com/user/logdrift/internal/differ"
)

// SummaryStats holds aggregate statistics for a diff session.
type SummaryStats struct {
	TotalLines   int
	Divergences  int
	Services     []string
	FieldChanges map[string]int // field name -> change count
}

// NewSummaryStats initialises a SummaryStats ready for accumulation.
func NewSummaryStats(services []string) *SummaryStats {
	return &SummaryStats{
		Services:     services,
		FieldChanges: make(map[string]int),
	}
}

// Record accumulates a slice of Deltas into the running stats.
func (s *SummaryStats) Record(deltas []differ.Delta) {
	s.TotalLines++
	if len(deltas) == 0 {
		return
	}
	s.Divergences++
	for _, d := range deltas {
		s.FieldChanges[d.Field]++
	}
}

// Render returns a human-readable summary block using the provided ColorScheme.
func (s *SummaryStats) Render(cs ColorScheme) string {
	var sb strings.Builder

	sb.WriteString(cs.Header("=== Session Summary ==="))
	sb.WriteByte('\n')
	sb.WriteString(fmt.Sprintf("  Services  : %s\n", strings.Join(s.Services, ", ")))
	sb.WriteString(fmt.Sprintf("  Lines     : %d\n", s.TotalLines))
	sb.WriteString(fmt.Sprintf("  Diverged  : %d\n", s.Divergences))

	if s.TotalLines > 0 {
		pct := float64(s.Divergences) / float64(s.TotalLines) * 100
		sb.WriteString(fmt.Sprintf("  Drift %%   : %.1f%%\n", pct))
	}

	if len(s.FieldChanges) > 0 {
		sb.WriteString("  Top fields:\n")
		for field, count := range s.FieldChanges {
			sb.WriteString(fmt.Sprintf("    %-20s %d\n", field, count))
		}
	}

	return sb.String()
}
