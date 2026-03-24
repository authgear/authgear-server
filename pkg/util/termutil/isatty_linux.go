//go:build (linux || aix || zos) && !appengine && !tinygo

package termutil

import (
	"golang.org/x/sys/unix"
)

// IsTerminal returns true if the file descriptor is a terminal.
// For example, IsTerminal(os.Stdout.Fd())
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS) // #nosec G115 -- Linux file descriptors are `int`; this cast is the syscall API boundary expected by x/sys/unix.
	return err == nil
}
