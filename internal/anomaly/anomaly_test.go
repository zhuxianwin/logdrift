package anomaly

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

func makeEntry(service string, fields map[string]interface{}) parser.Entry {
	return parser.Entry{
		Service:   service,
		Timestamp: time.Now(),
		Fields:    fields,
	}
}

func TestNew_EmptyFields_ReturnsError(t *testing.T) {
	_, err := New([]string{}, 3.0, 5)
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestNew_ZeroThreshold_ReturnsError(t *testing.T) {
	_, err := New([]string{"latency"}, 0, 5)
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestNew_LowMinSamples_ReturnsError(t *testing.T) {
	_, err := New([]string{"latency"}, 3.0, 1)
	if err == nil {
		t.Fatal("expected error for minSamples < 2")
	}
}

func TestNew_Valid_ReturnsDetector(t *testing.T) {
	d, err := New([]string{"latency"}, 3.0, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil detector")
	}
}

func TestObserve_BelowMinSamples_NoDeviations(t *testing.T) {
	d, _ := New([]string{"latency"}, 3.0, 10)
	for i := 0; i < 9; i++ {
		devs := d.Observe(makeEntry("svc", map[string]interface{}{"latency": float64(i)}))
		if len(devs) != 0 {
			t.Fatalf("expected no deviations before minSamples, got %d", len(devs))
		}
	}
}

func TestObserve_NormalValues_NoDeviations(t *testing.T) {
	d, _ := New([]string{"latency"}, 3.0, 5)
	// Feed stable values around 100ms
	for i := 0; i < 20; i++ {
		d.Observe(makeEntry("svc", map[string]interface{}{"latency": 100.0}))
	}
	devs := d.Observe(makeEntry("svc", map[string]interface{}{"latency": 101.0}))
	if len(devs) != 0 {
		t.Fatalf("expected no deviations for normal value, got %v", devs)
	}
}

func TestObserve_SpikeValue_ReturnsDeviation(t *testing.T) {
	d, _ := New([]string{"latency"}, 3.0, 5)
	// Establish a tight distribution around 10
	for i := 0; i < 30; i++ {
		d.Observe(makeEntry("svc", map[string]interface{}{"latency": 10.0}))
	}
	// Inject a massive outlier
	devs := d.Observe(makeEntry("svc", map[string]interface{}{"latency": 10000.0}))
	if len(devs) != 1 {
		t.Fatalf("expected 1 deviation, got %d", len(devs))
	}
	if devs[0].Field != "latency" {
		t.Errorf("expected field 'latency', got %q", devs[0].Field)
	}
	if devs[0].ZScore < 3.0 {
		t.Errorf("expected z-score >= 3.0, got %f", devs[0].ZScore)
	}
}

func TestObserve_NonNumericField_Ignored(t *testing.T) {
	d, _ := New([]string{"level"}, 3.0, 5)
	for i := 0; i < 10; i++ {
		devs := d.Observe(makeEntry("svc", map[string]interface{}{"level": "info"}))
		if len(devs) != 0 {
			t.Fatalf("expected no deviations for string field, got %d", len(devs))
		}
	}
}

func TestObserve_MissingField_Ignored(t *testing.T) {
	d, _ := New([]string{"latency"}, 3.0, 5)
	devs := d.Observe(makeEntry("svc", map[string]interface{}{"other": 42.0}))
	if len(devs) != 0 {
		t.Fatalf("expected no deviations for missing field, got %d", len(devs))
	}
}

func TestObserve_ServicesIsolated(t *testing.T) {
	d, _ := New([]string{"latency"}, 3.0, 5)
	// Seed svc-a with stable data
	for i := 0; i < 30; i++ {
		d.Observe(makeEntry("svc-a", map[string]interface{}{"latency": 10.0}))
	}
	// svc-b has no history yet — should not trigger anomaly
	devs := d.Observe(makeEntry("svc-b", map[string]interface{}{"latency": 10000.0}))
	if len(devs) != 0 {
		t.Fatalf("svc-b should be isolated from svc-a stats, got %d deviations", len(devs))
	}
}
