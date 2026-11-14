package curfew

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestNextWarning(t *testing.T) {
	clock := func(day, hour, minute int) time.Time {
		return time.Date(2025, 1, day, hour, minute, 0, 0, time.UTC)
	}
	tests := []struct {
		start string
		now   time.Time
		want  time.Time
	}{
		{"22:00", clock(1, 21, 44), clock(1, 21, 45)},
		{"22:00", clock(1, 21, 45), clock(1, 21, 45)},
		{"22:00", clock(1, 21, 46), clock(2, 21, 45)},
		{"00:10", clock(1, 23, 54), clock(1, 23, 55)},
		{"00:10", clock(1, 23, 56), clock(2, 23, 55)},
	}
	for _, tt := range tests {
		s, err := newService(Config{tt.start, "07:00", []string{"steam"}, 120})
		if err != nil {
			t.Fatal(err)
		}
		if got := s.nextWarning(tt.now); !got.Equal(tt.want) {
			t.Errorf("nextWarning(%v) = %v, want %v", tt.now, got, tt.want)
		}
	}
}

func TestWarningIndependentOfPolling(t *testing.T) {
	s, err := newService(Config{"22:00", "07:00", []string{"firefox", "steam"}, 120})
	if err != nil {
		t.Fatal(err)
	}
	s.find = func(apps []string) ([]int, error) {
		if len(apps) == 1 && apps[0] == "firefox" {
			return []int{101}, nil
		}
		return nil, nil
	}
	notified := make(chan []string, 1)
	s.warn = func(apps []string) error { notified <- apps; return nil }
	s.stop = func([]int) {}
	polls := make(chan time.Time)
	warnings := make(chan time.Time)
	reset := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.run(ctx, polls, warnings, func() { reset <- struct{}{} }) }()
	defer func() { cancel(); <-done }()

	// No poll is sent: the warning timer alone triggers the notification.
	warnings <- time.Now()
	select {
	case got := <-notified:
		if want := []string{"firefox"}; !reflect.DeepEqual(got, want) {
			t.Errorf("warning = %v, want %v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("warning not sent")
	}
	select {
	case <-reset:
	case <-time.After(time.Second):
		t.Fatal("warning timer not reset")
	}
}

func TestServiceChecks(t *testing.T) {
	s, err := newService(Config{"22:00", "07:00", []string{"steam"}, 1})
	if err != nil {
		t.Fatal(err)
	}
	s.find = func([]string) ([]int, error) { return []int{123}, nil }
	var stopped [][]int
	s.stop = func(pids []int) { stopped = append(stopped, pids) }
	for _, hour := range []int{21, 22, 3, 7} {
		if err := s.check(time.Date(2025, 1, 1, hour, 0, 0, 0, time.Local)); err != nil {
			t.Fatal(err)
		}
	}
	if want := [][]int{{123}, {123}}; !reflect.DeepEqual(stopped, want) {
		t.Errorf("stopped = %v, want %v", stopped, want)
	}
}

func TestNoWarningWhenNoneRunning(t *testing.T) {
	s, err := newService(Config{"00:10", "07:00", []string{"steam"}, 1})
	if err != nil {
		t.Fatal(err)
	}
	s.find = func([]string) ([]int, error) { return nil, nil }
	s.warn = func([]string) error { t.Error("unexpected warning"); return nil }
	if err := s.sendWarning(); err != nil {
		t.Fatal(err)
	}
}
