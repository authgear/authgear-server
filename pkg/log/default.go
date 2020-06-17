package log

import (
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

var defaultPatterns = []MaskPattern{
	// JWT
	NewRegexMaskPattern(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`),
	// Session Tokens
	NewRegexMaskPattern(`[A-Fa-f0-9-]{36}\.[A-Za-z0-9]+`),
}

func NewDefaultLogHook() logrus.Hook {
	patterns := defaultPatterns[:]

	return &FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}

func NewSecretLogHook(cfg *config.SecretConfig) logrus.Hook {
	var patterns []MaskPattern
	for _, item := range cfg.Secrets {
		for _, s := range item.Data.SensitiveStrings() {
			patterns = append(patterns, NewPlainMaskPattern(s))
		}
	}

	return &FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}
