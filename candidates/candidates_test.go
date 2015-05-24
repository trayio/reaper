package candidates

import (
	"testing"
	"time"
)

var now = time.Now().UTC()

var c = Candidates{
	{ID: "01", CreatedAt: now},
	{ID: "02", CreatedAt: now.Add(time.Hour * -3)},
	{ID: "03", CreatedAt: now.Add(time.Hour * -24)},
}

func TestOlderThanOneHour(t *testing.T) {

	x := c.OlderThan(1)
	if len(x) != 2 {
		t.Errorf("Expected 2 candidates, got %d\n", len(x))
	}

	if x[0].ID != "02" && x[1].ID != "03" {
		t.Errorf("Returned candidates are incorrect: %s\n", x)
	}
}

func TestOlderThanTenHours(t *testing.T) {
	x := c.OlderThan(10)
	if len(x) != 1 {
		t.Errorf("Expected 1 candidate, got %d\n", len(x))
	}

	if x[0].ID != "03" {
		t.Errorf("Returned candidates are incorrect: %s\n", x)
	}
}
