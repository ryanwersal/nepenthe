package consent

import (
	"fmt"
	"io"
	"os"

	"github.com/ryanwersal/nepenthe/internal/config"
	"golang.org/x/term"
)

var tierLabels = map[int]string{
	2: "Xcode / Apple Developer caches",
	3: "Docker containers and images",
}

func CheckTierConsent(tier int, cfg config.Config, w io.Writer) (bool, error) {
	var consent *bool
	switch tier {
	case 2:
		consent = cfg.Consent.Tier2
	case 3:
		consent = cfg.Consent.Tier3
	default:
		return true, nil
	}

	if consent != nil {
		return *consent, nil
	}

	// Non-interactive: auto-decline
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, nil
	}

	label := tierLabels[tier]
	fmt.Fprintf(w, "Tier %d: %s\nExclude these paths? [y/N] ", tier, label)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false, err
	}

	buf := make([]byte, 1)
	_, err = os.Stdin.Read(buf)

	// Restore terminal before handling error or printing
	term.Restore(int(os.Stdin.Fd()), oldState)

	if err != nil {
		return false, err
	}

	answer := buf[0] == 'y' || buf[0] == 'Y'

	if answer {
		fmt.Fprintf(w, "y\nTier %d enabled.\n", tier)
	} else {
		fmt.Fprintf(w, "n\nTier %d skipped.\n", tier)
	}

	if err := config.SaveConsent(tier, answer); err != nil {
		return answer, err
	}

	return answer, nil
}
