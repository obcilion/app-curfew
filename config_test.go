package curfew

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"start":"22:00","end":"07:00","applications":["firefox","steam"],"poll_interval_seconds":1}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Config{"22:00", "07:00", []string{"firefox", "steam"}, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadConfig() = %+v, want %+v", got, want)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid JSON", `{`},
		{"missing fields", `{}`},
		{"invalid start", `{"start":"24:00","end":"07:00","applications":["steam"],"poll_interval_seconds":1}`},
		{"invalid end format", `{"start":"22:00","end":"7:00","applications":["steam"],"poll_interval_seconds":1}`},
		{"no applications", `{"start":"22:00","end":"07:00","applications":[],"poll_interval_seconds":1}`},
		{"empty application", `{"start":"22:00","end":"07:00","applications":[""],"poll_interval_seconds":1}`},
		{"application path", `{"start":"22:00","end":"07:00","applications":["/usr/bin/steam"],"poll_interval_seconds":1}`},
		{"zero interval", `{"start":"22:00","end":"07:00","applications":["steam"],"poll_interval_seconds":0}`},
		{"negative interval", `{"start":"22:00","end":"07:00","applications":["steam"],"poll_interval_seconds":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadConfig(path); err == nil {
				t.Error("loadConfig() succeeded, want an error")
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	if _, err := loadConfig(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("loadConfig() succeeded, want an error")
	}
}
