package debug

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// PrintStack is equivalent to runtime/debug.PrintStack
func PrintStack() {
	os.Stderr.Write(Stack())
}

// Stack is equivalent to runtime/debug.Stack except that
// ALL gorountine stacks are included, not just the calling one.
func Stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

// TrapSIGQUIT traps SIGQUIT and call PrintStack.
// it DOES NOT exit the program.
func TrapSIGQUIT() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT)
	go func() {
		for range c {
			PrintStack()
		}
	}()
}
