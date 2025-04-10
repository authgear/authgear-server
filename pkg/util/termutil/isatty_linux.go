//go:build (linux || aix || zos) && !appengine && !tinygo

package termutil

import (
	"golang.org/x/sys/unix"
)

// IsTerminal returns true if the file descriptor is a terminal.
// For example, IsTerminal(os.Stdout.Fd())
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS)
	return err == nil
}
