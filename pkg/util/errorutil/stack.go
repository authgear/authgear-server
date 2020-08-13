package errorutil

import (
	"fmt"
	"runtime"
	"strings"
)

type Stacker interface {
	Stack() string
}

func formatFrame(frame runtime.Frame) string {
	fn := frame.Function
	if i := strings.LastIndex(fn, "/"); i >= 0 {
		fn = fn[i+1:]
	}
	return fmt.Sprintf("%s at %s:%d", fn, frame.File, frame.Line)
}

func stack(skip int) string {
	var rpc [2]uintptr
	n := runtime.Callers(skip+2, rpc[:])
	frame, ok := runtime.CallersFrames(rpc[:n]).Next()
	if !ok {
		return "<unknown>"
	}
	return formatFrame(frame)
}

func Callers(depth int) []string {
	pcs := make([]uintptr, depth)
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	st := make([]string, depth)
	n = 0
	for n < depth {
		frame, ok := frames.Next()
		if !ok {
			break
		}
		st[n] = formatFrame(frame)
		n++
	}
	return st[:n]
}
