package pipeline_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/pipeline"
)

type stringWriter struct {
	strings.Builder
}

func (sw *stringWriter) WriteString(s string) (int, error) {
	return sw.Builder.WriteString(s)
}

func writeLines(t *testing.T, path string, lines []string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer f.Close()
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
}

func TestRun_CancelImmediately(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.log")
	pathB := filepath.Join(dir, "b.log")
	writeLines(t, pathA, []string{})
	writeLines(t, pathB, []string{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := &stringWriter{}
	err := pipeline.Run(ctx, pipeline.Config{
		Services:   map[string]string{"svcA": pathA, "svcB": pathB},
		WindowSize: 10,
		Plain:      true,
		Output:     w,
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRun_EmitsDiffForDivergingLogs(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.log")
	pathB := filepath.Join(dir, "b.log")

	lineA := `{"time":"2024-01-01T00:00:00Z","level":"info","msg":"hello","status":200}`
	lineB := `{"time":"2024-01-01T00:00:00Z","level":"info","msg":"hello","status":500}`
	writeLines(t, pathA, []string{lineA})
	writeLines(t, pathB, []string{lineB})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w := &stringWriter{}
	done := make(chan error, 1)
	go func() {
		done <- pipeline.Run(ctx, pipeline.Config{
			Services:   map[string]string{"svcA": pathA, "svcB": pathB},
			WindowSize: 10,
			Plain:      true,
			Output:     w,
		})
	}()

	deadline := time.After(1500 * time.Millisecond)
	for {
		select {
		case <-deadline:
			cancel()
			<-done
			output := w.String()
			if !strings.Contains(output, "status") {
				t.Errorf("expected diff output to mention 'status', got:\n%s", output)
			}
			return
		case <-time.After(100 * time.Millisecond):
			if strings.Contains(w.String(), "status") {
				cancel()
				<-done
				return
			}
		}
	}
}
