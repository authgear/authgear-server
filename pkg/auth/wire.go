//+build wireinject

package auth

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

func NewAccessKeyMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(DependencySet, auth.ProvideAccessKeyMiddleware)
	return nil
}
