package curfew

import "time"

// inWindow reports whether now is in the daily window [start, end).
// Only the hour and minute of each time are used. Equal boundaries form an empty window.
func inWindow(now, start, end time.Time) bool {
	minutes := func(t time.Time) int {
		return t.Hour()*60 + t.Minute()
	}

	current := minutes(now)
	from := minutes(start)
	until := minutes(end)

	if from < until {
		return current >= from && current < until
	}
	if from > until {
		return current >= from || current < until
	}
	return false
}
