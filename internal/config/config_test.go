package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/logdrift/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "logdrift.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	p := writeConfig(t, `
services:
  - name: api
    path: /var/log/api.log
  - name: worker
    path: /var/log/worker.log
window_size: 50
output_mode: plain
`)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}
	if cfg.WindowSize != 50 {
		t.Errorf("expected window_size 50, got %d", cfg.WindowSize)
	}
	if cfg.OutputMode != "plain" {
		t.Errorf("expected output_mode plain, got %s", cfg.OutputMode)
	}
}

func TestLoad_Defaults(t *testing.T) {
	p := writeConfig(t, `
services:
  - name: svc
    path: /tmp/svc.log
`)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WindowSize != 100 {
		t.Errorf("expected default window_size 100, got %d", cfg.WindowSize)
	}
	if cfg.OutputMode != "color" {
		t.Errorf("expected default output_mode color, got %s", cfg.OutputMode)
	}
}

func TestLoad_NoServices(t *testing.T) {
	p := writeConfig(t, `services: []`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for empty services")
	}
}

func TestLoad_DuplicateServiceName(t *testing.T) {
	p := writeConfig(t, `
services:
  - name: api
    path: /a.log
  - name: api
    path: /b.log
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for duplicate service name")
	}
}

func TestLoad_InvalidOutputMode(t *testing.T) {
	p := writeConfig(t, `
services:
  - name: svc
    path: /tmp/svc.log
output_mode: json
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for invalid output_mode")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/logdrift.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidWindowSize(t *testing.T) {
	p := writeConfig(t, `
services:
  - name: svc
    path: /tmp/svc.log
window_size: -1
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for negative window_size")
	}
}
