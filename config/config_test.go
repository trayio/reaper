package config

import (
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	cfg := []byte(`
{
    "some": {
        "Count": 10,
        "Age": 24
    },

    "other": {
        "Count": 30,
        "Age": 72
    }
}
	`)

	c, err := generate(cfg)
	if err != nil {
		t.Errorf("Failed to create config: %s", err)
	}

	if len(c) != 2 {
		t.Errorf("Expected 2 elements, got %d\n", len(c))
	}

	if _, ok := c["some"]; !ok {
		t.Errorf("Missing 'some' key.")
	}

	if _, ok := c["other"]; !ok {
		t.Errorf("Missing 'other' key.")
	}

	if c["some"].Count != 10 {
		t.Errorf("Expected value 10 for attribute Count of key some, got %d\n", c["some"].Count)
	}

	if c["some"].Age != 24 {
		t.Errorf("Expected value 24 for attribute Age of key some, got %d\n", c["some"].Age)
	}

	if c["other"].Count != 30 {
		t.Errorf("Expected value 30 for attribute Count of key other, got %d\n", c["other"].Count)
	}

	if c["other"].Age != 72 {
		t.Errorf("Expected value 72 for attribute Age of key other, got %d\n", c["other"].Age)
	}
}
