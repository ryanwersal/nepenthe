package consent

import (
	"fmt"
	"io"
	"os"

	"github.com/ryanwersal/nepenthe/internal/config"
	"golang.org/x/term"
)

var categoryLabels = map[string]string{
	"dev-caches": "Xcode / Apple Developer caches",
	"containers": "Docker containers and images",
}

func CheckCategoryConsent(category string, cfg config.Config, w io.Writer) (bool, error) {
	label, needsConsent := categoryLabels[category]
	if !needsConsent {
		return true, nil
	}

	if cfg.Consent.Categories != nil {
		if answer, ok := cfg.Consent.Categories[category]; ok {
			return answer, nil
		}
	}

	// Non-interactive: auto-decline
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, nil
	}

	_, _ = fmt.Fprintf(w, "%s: %s\nExclude these paths? [y/N] ", category, label)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false, err
	}

	buf := make([]byte, 1)
	_, err = os.Stdin.Read(buf)

	// Restore terminal before handling error or printing
	_ = term.Restore(int(os.Stdin.Fd()), oldState)

	if err != nil {
		return false, err
	}

	answer := buf[0] == 'y' || buf[0] == 'Y'

	if answer {
		_, _ = fmt.Fprintf(w, "y\n%s enabled.\n", category)
	} else {
		_, _ = fmt.Fprintf(w, "n\n%s skipped.\n", category)
	}

	if err := config.SaveConsent(category, answer); err != nil {
		return answer, err
	}

	return answer, nil
}
