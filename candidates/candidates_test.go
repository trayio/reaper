package candidates

import (
	"sort"
	"testing"
	"time"
)

var now = time.Now().UTC()

var c = Candidates{
	{ID: "01", CreatedAt: now},
	{ID: "02", CreatedAt: now.Add(time.Hour * -3)},
	{ID: "03", CreatedAt: now.Add(time.Hour * -24)},
	{ID: "04", CreatedAt: now.Add(time.Minute * -119)},
	{ID: "05", CreatedAt: now.Add(time.Hour * -30)},
}

func TestOlderThanOneHour(t *testing.T) {

	x := c.OlderThan(1)
	if len(x) != 4 {
		t.Errorf("Expected 4 candidates, got %d\n", len(x))
	}

	if x[0].ID != "02" && x[1].ID != "03" && x[2].ID != "04" && x[3].ID != "05" {
		t.Errorf("Returned candidates are incorrect: %#v\n", x)
	}
}

func TestOlderThanTenHours(t *testing.T) {
	x := c.OlderThan(10)
	if len(x) != 2 {
		t.Errorf("Expected 2 candidates, got %d\n", len(x))
	}

	if x[0].ID != "03" && x[0].ID != "05" {
		t.Errorf("Returned candidates are incorrect: %#v\n", x)
	}
}

func TestOrder(t *testing.T) {
	x := c.OlderThan(1)
	sort.Sort(x)

	if x[0].ID != "04" {
		t.Errorf("First instance isn't closest to the next hour.")
	}
}
