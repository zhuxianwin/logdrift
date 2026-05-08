// Package routing provides field-based log routing to named output channels.
package routing

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/user/logdrift/internal/parser"
)

// Rule defines a single routing rule: if Field matches Pattern, send to Channel.
type Rule struct {
	Field   string
	Pattern string
	Channel string
	regexp  *regexp.Regexp
}

// Router routes parsed log entries to named channels based on field match rules.
type Router struct {
	mu       sync.RWMutex
	rules    []Rule
	channels map[string]chan parser.Entry
}

// New creates a Router from the provided rules.
// Returns an error if any rule has an invalid pattern or blank fields.
func New(rules []Rule) (*Router, error) {
	compiled := make([]Rule, 0, len(rules))
	for i, r := range rules {
		if r.Field == "" {
			return nil, fmt.Errorf("rule %d: field must not be empty", i)
		}
		if r.Channel == "" {
			return nil, fmt.Errorf("rule %d: channel must not be empty", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %d: invalid pattern %q: %w", i, r.Pattern, err)
		}
		r.regexp = re
		compiled = append(compiled, r)
	}
	return &Router{
		rules:    compiled,
		channels: make(map[string]chan parser.Entry),
	}, nil
}

// Channel returns (or lazily creates) a receive-only channel for the given name.
func (rt *Router) Channel(name string) <-chan parser.Entry {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if ch, ok := rt.channels[name]; ok {
		return ch
	}
	ch := make(chan parser.Entry, 64)
	rt.channels[name] = ch
	return ch
}

// Route evaluates entry against all rules and sends it to every matching channel.
// Rules are evaluated in order; an entry may match multiple rules.
func (rt *Router) Route(entry parser.Entry) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	for _, rule := range rt.rules {
		val, ok := entry.Fields[rule.Field]
		if !ok {
			continue
		}
		if rule.regexp.MatchString(fmt.Sprintf("%v", val)) {
			if ch, exists := rt.channels[rule.Channel]; exists {
				select {
				case ch <- entry:
				default:
				}
			}
		}
	}
}

// Channels returns the names of all registered channels.
func (rt *Router) Channels() []string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	names := make([]string, 0, len(rt.channels))
	for name := range rt.channels {
		names = append(names, name)
	}
	return names
}
