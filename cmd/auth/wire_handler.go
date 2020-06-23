//+build wireinject

package main

import (
	"net/http"

	"github.com/google/wire"

	handlersession "github.com/skygeario/skygear-server/pkg/auth/handler/session"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func newSessionResolveHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlersession.ResolveHandler)),
	))
}
