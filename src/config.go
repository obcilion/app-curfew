package curfew

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config describes a daily curfew. PollIntervalSeconds is the time between checks.
type Config struct {
	Start               string   `json:"start"`
	End                 string   `json:"end"`
	Applications        []string `json:"applications"`
	PollIntervalSeconds int      `json:"poll_interval_seconds"`
}

func loadConfig(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}
	if err := config.validate(); err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}

func (c Config) validate() error {
	for _, field := range []struct {
		name  string
		value string
	}{{"start", c.Start}, {"end", c.End}} {
		parsed, err := time.Parse("15:04", field.value)
		if err != nil || parsed.Format("15:04") != field.value {
			return fmt.Errorf("%s must be a time in HH:MM format", field.name)
		}
	}

	if len(c.Applications) == 0 {
		return fmt.Errorf("applications must not be empty")
	}
	for _, app := range c.Applications {
		if app == "" || strings.TrimSpace(app) != app || strings.Contains(app, "/") {
			return fmt.Errorf("invalid application name %q", app)
		}
	}
	if c.PollIntervalSeconds <= 0 {
		return fmt.Errorf("poll_interval_seconds must be positive")
	}
	return nil
}
