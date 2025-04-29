package termutil

import (
	"os"
)

// StdinStdoutIsTerminal checks if both stdin and stdout is a terminal.
// A typical terminal application requires both to be a terminal.
// Note that stderr is not checked because we allow the messages
// of the terminal program to be piped to somewhere else.
func StdinStdoutIsTerminal() bool {
	return IsTerminal(os.Stdin.Fd()) && IsTerminal(os.Stdout.Fd())
}
