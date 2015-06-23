package candidates

import (
	"time"
)

type Candidate struct {
	ID        string
	CreatedAt time.Time
}

type Candidates []Candidate

type Group map[string]Candidates

func (c Candidates) Len() int {
	return len(c)
}

// Sort by minute so the first in the list is the one closest to the next hour.
// Similar to ClosestToNextInstanceHour termination policy.
func (c Candidates) Less(i, j int) bool {
	return c[i].CreatedAt.Minute() > c[j].CreatedAt.Minute()
}

func (c Candidates) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Candidates) OlderThan(hours int) Candidates {
	candidates := Candidates{}

	now := time.Now().UTC()
	past := now.Add(time.Hour * time.Duration(-hours))

	for _, candidate := range c {
		if candidate.CreatedAt.Before(past) {
			candidates = append(candidates, candidate)
		}
	}

	return candidates
}
