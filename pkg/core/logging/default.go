package logging

var defaultPatterns = []MaskPattern{
	// JWT
	NewRegexMaskPattern(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`),
	// Session Tokens
	NewRegexMaskPattern(`[A-Fa-f0-9-]{36}\.[A-Za-z0-9]+`),
}

func NewDefaultMaskedTextFormatter(sensitiveStrings []string) *MaskedTextFormatter {
	patterns := defaultPatterns[:]
	if len(sensitiveStrings) != 0 {
		plainPatterns := make([]MaskPattern, len(sensitiveStrings))
		n := 0
		for _, s := range sensitiveStrings {
			if len(s) == 0 {
				continue
			}
			plainPatterns[n] = NewPlainMaskPattern(s)
			n++
		}
		patterns = append(patterns, plainPatterns[:n]...)
	}

	return &MaskedTextFormatter{
		Patterns: patterns,
		Mask:     "********",
	}
}
