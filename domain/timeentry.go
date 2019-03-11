package domain

import "time"

type (
	TimeEntry struct {
		aggregateRoot
		ID      string
		Project *Project
		Type    string
		Start   time.Time
		End     time.Time
		Breaks  time.Duration
		Comment string
	}
)
