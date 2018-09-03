package middleware

import "github.com/skygeario/skygear-server/pkg/core/handler"

type Middleware interface {
	Handle(handler.Handler) handler.Handler
}

func ApplyMiddlewares(
	h handler.Handler,
	middlewares ...Middleware,
) handler.Handler {
	for _, middleware := range middlewares {
		h = middleware.Handle(h)
	}
	return h
}
