package launchd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const label = "com.nepenthe.scan"

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func GeneratePlist(programArgs []string, intervalSeconds int) string {
	var args strings.Builder
	for _, arg := range programArgs {
		fmt.Fprintf(&args, "    <string>%s</string>\n", escapeXML(arg))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
%s  </array>
  <key>StartInterval</key>
  <integer>%d</integer>
  <key>RunAtLoad</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/tmp/nepenthe.stdout.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/nepenthe.stderr.log</string>
</dict>
</plist>
`, label, args.String(), intervalSeconds)
}

func Install(binaryPath string, intervalSeconds int) error {
	programArgs := []string{binaryPath, "scan", "--all", "--log-format", "json"}
	plist := GeneratePlist(programArgs, intervalSeconds)

	path, err := plistPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return err
	}

	// Use bootstrap (the modern replacement for deprecated "launchctl load")
	uid := fmt.Sprintf("gui/%d", os.Getuid())
	if err := exec.Command("launchctl", "bootstrap", uid, path).Run(); err != nil {
		return fmt.Errorf("launchctl bootstrap failed: %w", err)
	}

	return nil
}

func Uninstall() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("no plist found at %s", path)
	}

	// Use bootout (the modern replacement for deprecated "launchctl unload")
	// Ignore errors since the service may not be loaded
	uid := fmt.Sprintf("gui/%d", os.Getuid())
	_ = exec.Command("launchctl", "bootout", uid, path).Run()

	return os.Remove(path)
}

func PlistFilePath() (string, error) {
	return plistPath()
}
