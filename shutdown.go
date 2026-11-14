package curfew

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// stopProcess asks a process to exit, then kills it after the grace period.
func stopProcess(process *os.Process, gracePeriod time.Duration) error {
	if err := process.Signal(syscall.SIGTERM); err != nil {
		if processGone(err) {
			return nil
		}
		return fmt.Errorf("send SIGTERM: %w", err)
	}

	time.Sleep(gracePeriod)
	if err := process.Kill(); err != nil && !processGone(err) {
		return fmt.Errorf("send SIGKILL: %w", err)
	}
	return nil
}

func processGone(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
}
