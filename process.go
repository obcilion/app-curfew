package curfew

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// findProcesses returns the PIDs owned by uid whose executable names match apps.
// Processes that disappear or cannot be inspected during the scan are skipped.
func findProcesses(procDir string, uid int, apps []string) ([]int, error) {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil, fmt.Errorf("read process directory: %w", err)
	}

	var matches []int
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 || !entry.IsDir() {
			continue
		}

		dir := filepath.Join(procDir, entry.Name())
		status, err := os.ReadFile(filepath.Join(dir, "status"))
		if err != nil || !ownedBy(string(status), uid) {
			continue
		}
		executable, err := os.Readlink(filepath.Join(dir, "exe"))
		if err != nil {
			continue
		}
		for _, app := range apps {
			if filepath.Base(executable) == app {
				matches = append(matches, pid)
				break
			}
		}
	}
	return matches, nil
}

func ownedBy(status string, uid int) bool {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return false
			}
			owner, err := strconv.Atoi(fields[1])
			return err == nil && owner == uid
		}
	}
	return false
}
