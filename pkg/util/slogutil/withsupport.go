package slogutil

import (
	"log/slog"

	"github.com/jba/slog/withsupport"
)

func LinearizeGroupOrAttrs(groupOrAttrs *withsupport.GroupOrAttrs) []slog.Attr {
	attrs := []slog.Attr{}
	groupOrAttrs.Apply(func(groups []string, attr slog.Attr) {
		for i := len(groups) - 1; i >= 0; i = i - 1 {
			group := groups[i]
			attr = slog.Group(group, attr)
		}
		attrs = append(attrs, attr)
	})
	return attrs
}
