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

	// Copied from stdlib
	// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
	at := strings.LastIndex(a.Address, "@")
	var local, domain string
	if at < 0 {
		local = a.Address
	} else {
		local, domain = a.Address[:at], a.Address[at+1:]
	}
	// Copied from stdlib

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
