package mail

import (
	gomail "net/mail"
	"strings"
)

// MaskAddress masks the given email address, ignoring name.
func MaskAddress(s string) string {
	a, err := gomail.ParseAddress(s)
	if err != nil {
		return ""
	}

	local, domain := SplitAddress(a.Address)

	runes := []rune(local)
	length := len(runes)
	halfLength := length / 2

	var buf strings.Builder
	for i := 0; i < length; i++ {
		if i < halfLength {
			buf.WriteRune(runes[i])
		} else {
			buf.WriteRune('*')
		}
	}
	buf.WriteRune('@')
	buf.WriteString(domain)

	return buf.String()
}
