//+build wireinject

package challenge

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
)

func provideCreateHandler(h *CreateHandler) http.Handler {
	return h
}

func newCreateHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(challengeProvider), new(*challenge.Provider)),
		wire.Struct(new(CreateHandler), "*"),
		provideCreateHandler,
	)
	return nil
}
