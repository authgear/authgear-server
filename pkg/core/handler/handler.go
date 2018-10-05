package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, context.AuthContext)
}

type HandlerFunc func(http.ResponseWriter, *http.Request, context.AuthContext)

func (f HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.AuthContext) {
	f(rw, r, ctx)
}
