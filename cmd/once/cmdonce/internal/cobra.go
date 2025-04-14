package internal

import (
	"fmt"
	"os"
)

// PrintError prints err to stderr and returns err.
func PrintError(err error) error {
	fatalError := FatalError{Err: err}
	fmt.Fprintf(os.Stderr, "%v", fatalError.View())
	return err
}
