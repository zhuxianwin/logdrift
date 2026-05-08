// Package anomaly detects statistical anomalies in log field values
// by tracking rolling mean and standard deviation per field.
package anomaly

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

// Deviation describes a single anomalous field value.
type Deviation struct {
	Field   string
	Value   float64
	Mean    float64
	StdDev  float64
	ZScore  float64
	Service string
	At      time.Time
}

type stats struct {
	n    float64
	mean float64
	m2   float64 // Welford's online algorithm accumulator
}

func (s *stats) update(x float64) {
	s.n++
	delta := x - s.mean
	s.mean += delta / s.n
	delta2 := x - s.mean
	s.m2 += delta * delta2
}

func (s *stats) stddev() float64 {
	if s.n < 2 {
		return 0
	}
	return math.Sqrt(s.m2 / (s.n - 1))
}

// Detector tracks per-service, per-field numeric statistics and flags
// values whose z-score exceeds the configured threshold.
type Detector struct {
	mu        sync.Mutex
	threshold float64
	minSamples int
	fields    []string
	state     map[string]*stats // key: "service:field"
}

// New creates a Detector. threshold is the z-score cutoff (e.g. 3.0).
// minSamples is the minimum number of observations before anomalies are reported.
func New(fields []string, threshold float64, minSamples int) (*Detector, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("anomaly: at least one field required")
	}
	if threshold <= 0 {
		return nil, fmt.Errorf("anomaly: threshold must be positive, got %f", threshold)
	}
	if minSamples < 2 {
		return nil, fmt.Errorf("anomaly: minSamples must be >= 2, got %d", minSamples)
	}
	return &Detector{
		threshold:  threshold,
		minSamples: minSamples,
		fields:     fields,
		state:      make(map[string]*stats),
	}, nil
}

// Observe records field values from entry and returns any detected deviations.
func (d *Detector) Observe(entry parser.Entry) []Deviation {
	d.mu.Lock()
	defer d.mu.Unlock()

	var deviations []Deviation
	for _, field := range d.fields {
		raw, ok := entry.Fields[field]
		if !ok {
			continue
		}
		var v float64
		switch val := raw.(type) {
		case float64:
			v = val
		case int:
			v = float64(val)
		default:
			continue
		}

		key := entry.Service + ":" + field
		st, exists := d.state[key]
		if !exists {
			st = &stats{}
			d.state[key] = st
		}

		st.update(v)

		if int(st.n) < d.minSamples {
			continue
		}
		sd := st.stddev()
		if sd == 0 {
			continue
		}
		z := math.Abs(v-st.mean) / sd
		if z >= d.threshold {
			deviations = append(deviations, Deviation{
				Field:   field,
				Value:   v,
				Mean:    st.mean,
				StdDev:  sd,
				ZScore:  z,
				Service: entry.Service,
				At:      entry.Timestamp,
			})
		}
	}
	return deviations
}
