package handler

import (
	"net/http"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, AuthenticationContext)
}

type HandlerFunc func(http.ResponseWriter, *http.Request, AuthenticationContext)

func (f HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx AuthenticationContext) {
	f(rw, r, ctx)
}
