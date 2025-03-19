package main

import (
	"context"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/authgear/authgear-server/cmd/once/cmdonce"
	"github.com/authgear/authgear-server/pkg/util/debug"
)

func main() {
	_, _ = maxprocs.Set()

	debug.TrapSIGQUIT()

	// This program does not load .env at the moment.

	ctx := context.Background()
	cmdonce.Run(ctx)
}
