package mail

import (
	"strings"
)

// SplitAddress splits s into local part and domain.
func SplitAddress(s string) (local string, domain string) {
	// Copied from stdlib
	// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
	at := strings.LastIndex(s, "@")
	if at < 0 {
		local = s
		domain = ""
	} else {
		local, domain = s[:at], s[at+1:]
	}
	// Copied from stdlib

	return
}
