package curfew

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestShutdownHelper(t *testing.T) {
	mode := os.Getenv("CURFEW_TEST_HELPER")
	if mode == "" {
		return
	}
	if mode == "ignore" {
		signal.Ignore(syscall.SIGTERM)
	}
	fmt.Println("ready")
	time.Sleep(time.Minute)
	os.Exit(0)
}

func TestStopProcess(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mode   string
		signal syscall.Signal
	}{
		{"terminates normally", "normal", syscall.SIGTERM},
		{"kills process ignoring SIGTERM", "ignore", syscall.SIGKILL},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestShutdownHelper$")
			cmd.Env = append(os.Environ(), "CURFEW_TEST_HELPER="+tt.mode)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = cmd.Process.Kill() })
			if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
				t.Fatalf("helper readiness: %q, %v", line, err)
			}

			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			if err := stopProcess(cmd.Process, 100*time.Millisecond); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatalf("Wait() = %v, want signal %v", err, tt.signal)
				}
				status := exitErr.Sys().(syscall.WaitStatus)
				if !status.Signaled() || status.Signal() != tt.signal {
					t.Errorf("exit status = %v, want signal %v", status, tt.signal)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("process did not exit")
			}
		})
	}
}
