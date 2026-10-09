package attrs

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// ValidateStringConstraints emits one error per failed constraint on vctx.
// An empty value is not checked.
func ValidateStringConstraints(vctx *validation.Context, c *config.StringConstraints, value string, multiline bool) {
	if c == nil || value == "" {
		return
	}

	if c.MinLength != nil {
		if n := utf8.RuneCountInString(value); n < *c.MinLength {
			vctx.EmitError("minLength", map[string]any{"expected": *c.MinLength, "actual": n})
		}
	}

	if c.MaxLength != nil {
		if n := utf8.RuneCountInString(value); n > *c.MaxLength {
			vctx.EmitError("maxLength", map[string]any{"expected": *c.MaxLength, "actual": n})
		}
	}

	if len(c.AllowedCharacters) > 0 && !allCharactersAllowed(c.AllowedCharacters, value, multiline) {
		vctx.EmitError("allowedCharacters", map[string]any{"expected": c.AllowedCharacters})
	}

	if c.Pattern != "" {
		// The pattern was checked by the x_re2_regex format when the config was loaded.
		// It is compiled on each check because a process-wide cache would grow with every tenant and config edit.
		if !regexp.MustCompile(c.Pattern).MatchString(value) {
			vctx.EmitError("pattern", map[string]any{"expected": c.Pattern})
		}
	}
}

func allCharactersAllowed(allowed []config.AllowedCharacters, value string, multiline bool) bool {
	// The kind of the previous character, which a mark or a joiner takes.
	var prevKind config.AllowedCharacters
	for _, r := range value {
		if multiline && (r == '\n' || r == '\r') {
			prevKind = ""
			continue
		}
		var kind config.AllowedCharacters
		var ok bool
		// A mark, such as an accent or an emoji variation selector, or a joiner, belongs to the character before it.
		if unicode.Is(unicode.M, r) || isJoiner(r) {
			kind, ok = prevKind, prevKind != ""
		} else {
			kind, ok = runeKind(r)
		}
		prevKind = kind
		if !ok {
			return false
		}
		found := false
		for _, a := range allowed {
			if a == kind {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// runeKind returns false for runes that belong to no kind, such as control characters and tabs.
func runeKind(r rune) (config.AllowedCharacters, bool) {
	switch {
	case unicode.IsLetter(r):
		return config.AllowedCharactersLetters, true
	case '0' <= r && r <= '9':
		return config.AllowedCharactersDigits, true
	case r == ' ':
		return config.AllowedCharactersSpaces, true
	case unicode.IsGraphic(r):
		return config.AllowedCharactersSymbols, true
	default:
		return "", false
	}
}

// isJoiner reports whether r is U+200C ZERO WIDTH NON-JOINER or U+200D ZERO WIDTH JOINER,
// which are part of words in scripts such as Persian, and of emoji sequences.
func isJoiner(r rune) bool {
	return r == '\u200c' || r == '\u200d'
}

// IsMultilineStandardAttribute reports whether pointer is /address/formatted or /address/street_address.
func IsMultilineStandardAttribute(pointer string) bool {
	return pointer == "/address/formatted" || pointer == "/address/street_address"
}
