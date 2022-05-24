//go:build authgearlite
// +build authgearlite

package vipsutil

import (
	"fmt"
	"io"
)

type Daemon struct{}

var _ io.Closer = &Daemon{}

var errPanic = fmt.Errorf("govips is not available in lite build")

func OpenDaemon(numWorker int) *Daemon {
	panic(errPanic)
}

func (v *Daemon) Close() error {
	panic(errPanic)
}

func (v *Daemon) Process(i Input) (*Output, error) {
	panic(errPanic)
}
