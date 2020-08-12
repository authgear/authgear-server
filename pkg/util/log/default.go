package log

import (
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var defaultPatterns = []MaskPattern{
	// JWT
	NewRegexMaskPattern(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`),
	// Session Tokens
	NewRegexMaskPattern(`[A-Fa-f0-9-]{36}\.[A-Za-z0-9]+`),
}

func NewDefaultMaskLogHook() logrus.Hook {
	patterns := defaultPatterns[:]

	return &FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}

func NewSecretMaskLogHook(cfg *config.SecretConfig) logrus.Hook {
	var patterns []MaskPattern
	for _, item := range cfg.Secrets {
		for _, s := range item.Data.SensitiveStrings() {
			if len(s) == 0 {
				continue
			}
			patterns = append(patterns, NewPlainMaskPattern(s))
		}
	}

	return &FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}
