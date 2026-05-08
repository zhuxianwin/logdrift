// Package enricher provides field enrichment for structured log entries.
//
// An Enricher holds a list of Rules. Each Rule reads a value from an existing
// field in a parser.Entry, optionally transforms it (uppercase, prefix), and
// writes the result into a new destination field.
//
// Rules never overwrite existing destination fields, and missing source fields
// are silently skipped so that enrichment is always non-destructive.
//
// Typical usage:
//
//	e, err := enricher.New([]enricher.Rule{
//		{SourceField: "env", DestField: "env_tag", Prefix: "env:"},
//		{SourceField: "level", DestField: "level_upper", Uppercase: true},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	enriched := e.Apply(entry)
//
// Configuration can also be loaded from a ServiceEnricherConfig via
// enricher.FromConfig, which integrates with the top-level config package.
package enricher
