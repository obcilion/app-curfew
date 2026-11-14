package curfew

import (
	"testing"
	"time"
)

func TestInWindow(t *testing.T) {
	clock := func(hour, minute int) time.Time {
		return time.Date(2025, time.January, 1, hour, minute, 0, 0, time.Local)
	}

	tests := []struct {
		name       string
		now        time.Time
		start, end time.Time
		want       bool
	}{
		{"daytime before", clock(8, 59), clock(9, 0), clock(17, 0), false},
		{"daytime start", clock(9, 0), clock(9, 0), clock(17, 0), true},
		{"daytime inside", clock(12, 0), clock(9, 0), clock(17, 0), true},
		{"daytime end", clock(17, 0), clock(9, 0), clock(17, 0), false},
		{"overnight before", clock(21, 59), clock(22, 0), clock(7, 0), false},
		{"overnight start", clock(22, 0), clock(22, 0), clock(7, 0), true},
		{"overnight midnight", clock(0, 0), clock(22, 0), clock(7, 0), true},
		{"overnight before end", clock(6, 59), clock(22, 0), clock(7, 0), true},
		{"overnight end", clock(7, 0), clock(22, 0), clock(7, 0), false},
		{"equal boundaries", clock(9, 0), clock(9, 0), clock(9, 0), false},
		{"equal boundaries later", clock(10, 0), clock(9, 0), clock(9, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inWindow(tt.now, tt.start, tt.end); got != tt.want {
				t.Errorf("inWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}
