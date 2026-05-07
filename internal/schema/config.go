package schema

import (
	"fmt"

	"github.com/robinmuhia/logdrift/internal/config"
)

// ServiceSchema holds a named schema collector per service.
type ServiceSchema struct {
	schemas map[string]*Schema
}

// BuildFromConfig creates a ServiceSchema pre-populated with a Schema for each
// service defined in cfg. Returns an error if cfg has no services.
func BuildFromConfig(cfg *config.Config) (*ServiceSchema, error) {
	if cfg == nil || len(cfg.Services) == 0 {
		return nil, fmt.Errorf("schema: at least one service is required")
	}
	ss := &ServiceSchema{
		schemas: make(map[string]*Schema, len(cfg.Services)),
	}
	for _, svc := range cfg.Services {
		ss.schemas[svc.Name] = New()
	}
	return ss, nil
}

// Get returns the Schema for the named service, and whether it exists.
func (ss *ServiceSchema) Get(service string) (*Schema, bool) {
	s, ok := ss.schemas[service]
	return s, ok
}

// Services returns the list of service names tracked by this ServiceSchema.
func (ss *ServiceSchema) Services() []string {
	names := make([]string, 0, len(ss.schemas))
	for name := range ss.schemas {
		names = append(names, name)
	}
	return names
}
