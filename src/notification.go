package curfew

import (
	"fmt"
	"os/exec"
	"strings"
)

// notify warns that the listed applications will close in 15 minutes.
func notify(apps []string) error {
	message := "These applications will close in 15 minutes:\n" + strings.Join(apps, "\n")
	if err := exec.Command("notify-send", "-u", "critical", "App curfew", message).Run(); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	return nil
}
