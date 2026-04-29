package tailer_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/tailer"
)

func TestManager_NoServices(t *testing.T) {
	m := tailer.NewManager(nil)
	ctx := context.Background()
	if err := m.Run(ctx); err == nil {
		t.Error("expected error for empty service list")
	}
}

func TestManager_FansOutLines(t *testing.T) {
	dir := t.TempDir()

	fileA, _ := os.CreateTemp(dir, "svc-a-*.log")
	fileB, _ := os.CreateTemp(dir, "svc-b-*.log")
	defer fileA.Close()
	defer fileB.Close()

	svcs := []tailer.ServiceConfig{
		{Name: "svc-a", Path: fileA.Name()},
		{Name: "svc-b", Path: fileB.Name()},
	}

	m := tailer.NewManager(svcs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { _ = m.Run(ctx) }()

	time.Sleep(60 * time.Millisecond)

	_, _ = fileA.WriteString(`{"level":"info","svc":"a"}` + "\n")
	_, _ = fileB.WriteString(`{"level":"warn","svc":"b"}` + "\n")

	received := map[string]bool{}
	timeout := time.After(2 * time.Second)
	for len(received) < 2 {
		select {
		case line := <-m.Out:
			received[line.Source] = true
		case <-timeout:
			t.Fatalf("timed out; only received from: %v", received)
		}
	}

	if !received["svc-a"] || !received["svc-b"] {
		t.Errorf("did not receive lines from all services: %v", received)
	}
}
