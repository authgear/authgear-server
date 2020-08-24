package log

import (
	"github.com/sirupsen/logrus"
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
