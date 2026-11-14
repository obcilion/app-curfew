package curfew

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotify(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "arguments")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$NOTIFY_TEST_OUTPUT\"\n"
	if err := os.WriteFile(filepath.Join(dir, "notify-send"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("NOTIFY_TEST_OUTPUT", output)

	if err := notify([]string{"firefox", "steam"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := "-u\ncritical\nApp curfew\nThese applications will close in 15 minutes:\nfirefox\nsteam\n"
	if string(got) != want {
		t.Errorf("notification arguments = %q, want %q", got, want)
	}
}

func TestNotifyMissingCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := notify([]string{"firefox"}); err == nil {
		t.Error("notify() succeeded, want an error")
	}
}
