package tmutil

import (
	"errors"
	"os/exec"
	"strings"
)

var ErrNotFound = errors.New("tmutil not found — is this macOS?")

func AssertAvailable() error {
	_, err := exec.LookPath("tmutil")
	if err != nil {
		return ErrNotFound
	}
	return nil
}

func IsExcluded(path string) (bool, error) {
	out, err := exec.Command("tmutil", "isexcluded", path).CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), "[Excluded]"), nil
}

func AddExclusion(path string) error {
	return exec.Command("tmutil", "addexclusion", path).Run()
}

func RemoveExclusion(path string) error {
	return exec.Command("tmutil", "removeexclusion", path).Run()
}

func ListAllExclusions() ([]string, error) {
	out, err := exec.Command("mdfind", "com_apple_backup_excludeItem = 'com.apple.backupd'").CombinedOutput()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}
