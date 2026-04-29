package tailer

import (
	"context"
	"fmt"
	"sync"
)

// ServiceConfig maps a human-readable label to a log file path.
type ServiceConfig struct {
	Name string
	Path string
}

// Manager owns multiple Tailers and fans their output into a single channel.
type Manager struct {
	services []ServiceConfig
	Out      chan Line
}

// NewManager creates a Manager for the given service configs.
func NewManager(services []ServiceConfig) *Manager {
	return &Manager{
		services: services,
		Out:      make(chan Line, 256),
	}
}

// Run starts a Tailer for every service and blocks until all finish or ctx is cancelled.
// Any tailer error is printed to stderr but does not stop the others.
func (m *Manager) Run(ctx context.Context) error {
	if len(m.services) == 0 {
		return fmt.Errorf("no services configured")
	}

	var wg sync.WaitGroup
	for _, svc := range m.services {
		wg.Add(1)
		go func(s ServiceConfig) {
			defer wg.Done()
			t := New(s.Name, s.Path, m.Out)
			if err := t.Tail(ctx); err != nil && err != ctx.Err() {
				fmt.Printf("[logdrift] tailer error for %s: %v\n", s.Name, err)
			}
		}(svc)
	}

	wg.Wait()
	close(m.Out)
	return nil
}
