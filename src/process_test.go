package curfew

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindProcesses(t *testing.T) {
	proc := t.TempDir()
	processes := []struct {
		pid, status, executable string
	}{
		{"101", "Name:\tfirefox\nUid:\t1000\t1000\t1000\t1000\n", "/usr/bin/firefox"},
		{"102", "Uid:\t1000\t1000\t1000\t1000\n", "/opt/firefox"},
		{"103", "Uid:\t1001\t1001\t1001\t1001\n", "/usr/bin/firefox"},
		{"104", "Uid:\t1000\t1000\t1000\t1000\n", "/usr/bin/steam"},
		{"105", "Uid:\tbad\n", "/usr/bin/firefox"},
		{"106", "Uid:\t1000\n", ""},
	}
	for _, process := range processes {
		dir := filepath.Join(proc, process.pid)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "status"), []byte(process.status), 0600); err != nil {
			t.Fatal(err)
		}
		if process.executable != "" {
			if err := os.Symlink(process.executable, filepath.Join(dir, "exe")); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.Mkdir(filepath.Join(proc, "not-a-pid"), 0700); err != nil {
		t.Fatal(err)
	}

	got, err := findProcesses(proc, 1000, []string{"firefox", "firefox"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{101, 102}; !reflect.DeepEqual(got, want) {
		t.Errorf("findProcesses() = %v, want %v", got, want)
	}
}

func TestFindProcessesMissingDirectory(t *testing.T) {
	_, err := findProcesses(filepath.Join(t.TempDir(), "missing"), 1000, []string{"firefox"})
	if err == nil {
		t.Error("findProcesses() succeeded, want an error")
	}
}
