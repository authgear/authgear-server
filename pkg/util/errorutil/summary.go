package errorutil

import (
	"strings"
)

func Summary(err error) string {
	var msgs []string
	lastMsg := ""

	Unwrap(err, func(err error) {
		var msg string
		if s, ok := err.(interface{ Summary() string }); ok {
			msg = s.Summary()
		} else {
			msg = err.Error()
		}

		if !strings.HasSuffix(lastMsg, msg) {
			msgs = append(msgs, msg)
			lastMsg = msg
		}
	})

	return strings.Join(msgs, ": ")
}
