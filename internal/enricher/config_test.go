package enricher

import "testing"

func TestFromConfig_Nil_ReturnsNil(t *testing.T) {
	e, err := FromConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e != nil {
		t.Fatal("expected nil enricher for nil config")
	}
}

func TestFromConfig_EmptyRules_ReturnsNil(t *testing.T) {
	e, err := FromConfig(&ServiceEnricherConfig{Rules: []RuleConfig{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e != nil {
		t.Fatal("expected nil enricher for empty rules")
	}
}

func TestFromConfig_ValidRules_ReturnsEnricher(t *testing.T) {
	cfg := &ServiceEnricherConfig{
		Rules: []RuleConfig{
			{SourceField: "env", DestField: "env_tag"},
		},
	}
	e, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil enricher")
	}
}

func TestFromConfig_EmptySourceField_ReturnsError(t *testing.T) {
	cfg := &ServiceEnricherConfig{
		Rules: []RuleConfig{
			{SourceField: "", DestField: "dst"},
		},
	}
	_, err := FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty source_field")
	}
}

func TestFromConfig_EmptyDestField_ReturnsError(t *testing.T) {
	cfg := &ServiceEnricherConfig{
		Rules: []RuleConfig{
			{SourceField: "src", DestField: ""},
		},
	}
	_, err := FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty dest_field")
	}
}

func TestFromConfig_PrefixAndUppercase_Propagated(t *testing.T) {
	cfg := &ServiceEnricherConfig{
		Rules: []RuleConfig{
			{SourceField: "region", DestField: "region_tag", Prefix: "r:", Uppercase: true},
		},
	}
	e, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.rules[0].Prefix != "r:" || !e.rules[0].Uppercase {
		t.Fatal("prefix or uppercase not propagated")
	}
}
