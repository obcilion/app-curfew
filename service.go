package curfew

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type service struct {
	config     Config
	start, end time.Time
	find       func([]string) ([]int, error)
	warn       func([]string) error
	stop       func([]int)
}

func newService(config Config) (*service, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	start, _ := time.Parse("15:04", config.Start)
	end, _ := time.Parse("15:04", config.End)

	var stopping sync.Map
	return &service{
		config: config,
		start:  start,
		end:    end,
		find: func(apps []string) ([]int, error) {
			return findProcesses("/proc", os.Getuid(), apps)
		},
		warn: notify,
		stop: func(pids []int) {
			for _, pid := range pids {
				if _, loaded := stopping.LoadOrStore(pid, struct{}{}); loaded {
					continue
				}
				go func(pid int) {
					defer stopping.Delete(pid)
					process, err := os.FindProcess(pid)
					if err == nil {
						err = stopProcess(process, 5*time.Second)
						_ = process.Release()
					}
					if err != nil {
						log.Printf("stop process %d: %v", pid, err)
					}
				}(pid)
			}
		},
	}, nil
}

// nextWarning returns the next warning time, including one that is due now.
func (s *service) nextWarning(now time.Time) time.Time {
	for day := 0; ; day++ {
		start := time.Date(now.Year(), now.Month(), now.Day()+day, s.start.Hour(), s.start.Minute(), 0, 0, now.Location())
		warning := start.Add(-15 * time.Minute)
		if !warning.Before(now) {
			return warning
		}
	}
}

func (s *service) sendWarning() error {
	var running []string
	for _, app := range s.config.Applications {
		pids, err := s.find([]string{app})
		if err != nil {
			return fmt.Errorf("find applications for warning: %w", err)
		}
		if len(pids) > 0 {
			running = append(running, app)
		}
	}
	if len(running) > 0 {
		return s.warn(running)
	}
	return nil
}

// check enforces the curfew at the given local time.
func (s *service) check(now time.Time) error {
	if inWindow(now, s.start, s.end) {
		pids, err := s.find(s.config.Applications)
		if err != nil {
			return fmt.Errorf("find applications to stop: %w", err)
		}
		s.stop(pids)
	}
	return nil
}

// runService loads the configuration and checks immediately and at each interval.
func runService(ctx context.Context, configPath string) error {
	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	s, err := newService(config)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(config.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	var warningTimer *time.Timer
	var warnings <-chan time.Time
	resetWarning := func() {}
	if config.Start != config.End {
		warningTimer = time.NewTimer(time.Until(s.nextWarning(time.Now())))
		defer warningTimer.Stop()
		warnings = warningTimer.C
		resetWarning = func() {
			warningTimer.Reset(time.Until(s.nextWarning(time.Now().Add(time.Nanosecond))))
		}
	}
	return s.run(ctx, ticker.C, warnings, resetWarning)
}

func (s *service) run(ctx context.Context, polls, warnings <-chan time.Time, resetWarning func()) error {
	if err := s.check(time.Now()); err != nil {
		log.Printf("curfew check: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-polls:
			if err := s.check(time.Now()); err != nil {
				log.Printf("curfew check: %v", err)
			}
		case <-warnings:
			if err := s.sendWarning(); err != nil {
				log.Printf("curfew warning: %v", err)
			}
			resetWarning()
		}
	}
}
