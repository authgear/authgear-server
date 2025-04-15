//go:build (darwin || freebsd || openbsd || netbsd || dragonfly || hurd) && !appengine && !tinygo

package termutil

import (
	"golang.org/x/sys/unix"
)

// IsTerminal returns true if the file descriptor is a terminal.
// For example, IsTerminal(os.Stdout.Fd())
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TIOCGETA)
	return err == nil
}
