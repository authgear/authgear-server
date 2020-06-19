package middlewares

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/deps"
)

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

func NewSessionMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*auth.Middleware)(nil).Handle
}

func NewCSPMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.CSPMiddleware)(nil).Handle
}

func NewCSRFMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.CSRFMiddleware)(nil).Handle
}

func NewStateMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.StateMiddleware)(nil).Handle
}

func NewClientIDMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.ClientIDMiddleware)(nil).Handle
}
