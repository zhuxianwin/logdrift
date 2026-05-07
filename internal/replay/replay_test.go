package replay

import (
	"context"
	"strings"
	"testing"
	"time"
)

const (
	line1 = `{"level":"info","msg":"start","time":"2024-01-01T00:00:00Z"}`
	line2 = `{"level":"info","msg":"middle","time":"2024-01-01T00:00:01Z"}`
	line3 = `{"level":"info","msg":"end","time":"2024-01-01T00:00:02Z"}`
)

func TestRun_EmitsAllLines(t *testing.T) {
	input := strings.Join([]string{line1, line2, line3}, "\n")
	ctx := context.Background()

	ch, err := Run(ctx, "svc", strings.NewReader(input), Options{NoDelay: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []Line
	for l := range ch {
		got = append(got, l)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	for _, l := range got {
		if l.Service != "svc" {
			t.Errorf("expected service 'svc', got %q", l.Service)
		}
	}
}

func TestRun_SkipsNonJSON(t *testing.T) {
	input := "not json\n" + line1
	ctx := context.Background()

	ch, err := Run(ctx, "svc", strings.NewReader(input), Options{NoDelay: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []Line
	for l := range ch {
		got = append(got, l)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %d", len(got))
	}
}

func TestRun_CancelStopsReplay(t *testing.T) {
	// Build a large input so the goroutine would block without cancel.
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(line1 + "\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	ch, err := Run(ctx, "svc", strings.NewReader(sb.String()), Options{NoDelay: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drain; should complete quickly.
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("replay did not stop after context cancellation")
	}
}

func TestRun_NoDelayIgnoresTiming(t *testing.T) {
	input := strings.Join([]string{line1, line2, line3}, "\n")
	ctx := context.Background()

	start := time.Now()
	ch, _ := Run(ctx, "svc", strings.NewReader(input), Options{NoDelay: true})
	for range ch {
	}
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Errorf("NoDelay replay took too long: %v", elapsed)
	}
}
