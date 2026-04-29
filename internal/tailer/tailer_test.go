package tailer_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/tailer"
)

func TestTailer_EmitsLines(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "logdrift-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	out := make(chan tailer.Line, 10)
	tlr := tailer.New("svc-a", tmpFile.Name(), out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = tlr.Tail(ctx)
	}()

	// Give the tailer time to seek to end before writing.
	time.Sleep(50 * time.Millisecond)

	_, err = tmpFile.WriteString(`{"level":"info","msg":"hello"}` + "\n")
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	select {
	case line := <-out:
		if line.Source != "svc-a" {
			t.Errorf("expected source svc-a, got %s", line.Source)
		}
		if line.Content == "" {
			t.Error("expected non-empty content")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for line")
	}
}

func TestTailer_CancelStops(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "logdrift-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	out := make(chan tailer.Line, 10)
	tlr := tailer.New("svc-b", tmpFile.Name(), out)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- tlr.Tail(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for tailer to stop")
	}
}
