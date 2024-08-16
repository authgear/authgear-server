package ldap

import (
	"encoding/hex"
	"strings"
)

func EncodeAttributeName(name string) string {
	return encodeString(name, false)
}

func EncodeAttributeValue(value string) string {
	return encodeString(value, true)
}

// Reference: https://github.com/go-ldap/ldap/blob/06d50d1ad03bcd323e48f2fe174d95ceb31b8b90/v3/dn.go#L172
// Escape a string according to RFC 4514
func encodeString(value string, isValue bool) string {
	builder := strings.Builder{}

	escapeChar := func(c byte) {
		builder.WriteByte('\\')
		builder.WriteByte(c)
	}

	escapeHex := func(c byte) {
		builder.WriteByte('\\')
		builder.WriteString(hex.EncodeToString([]byte{c}))
	}

	// Loop through each byte and escape as necessary.
	// Runes that take up more than one byte are escaped
	// byte by byte (since both bytes are non-ASCII).
	for i := 0; i < len(value); i++ {
		char := value[i]
		if i == 0 && (char == ' ' || char == '#') {
			// Special case leading space or number sign.
			escapeChar(char)
			continue
		}
		if i == len(value)-1 && char == ' ' {
			// Special case trailing space.
			escapeChar(char)
			continue
		}

		switch char {
		case '"', '+', ',', ';', '<', '>', '\\':
			// Each of these special characters must be escaped.
			escapeChar(char)
			continue
		}

		if !isValue && char == '=' {
			// Equal signs have to be escaped only in the type part of
			// the attribute type and value pair.
			escapeChar(char)
			continue
		}

		if char < ' ' || char > '~' {
			// All special character escapes are handled first
			// above. All bytes less than ASCII SPACE and all bytes
			// greater than ASCII TILDE must be hex-escaped.
			escapeHex(char)
			continue
		}

		// Any other character does not require escaping.
		builder.WriteByte(char)
	}

	return builder.String()
}
