package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, auth.AuthInfo)
}

type HandlerFunc func(http.ResponseWriter, *http.Request, auth.AuthInfo)

func (f HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	f(rw, r, authInfo)
}
