package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type PolicyProvider interface {
	ProvideAuthzPolicy() Policy
}

type Policy interface {
	IsAllowed(r *http.Request, ctx handler.AuthContext) error
}

type PolicyFunc func(r *http.Request, ctx handler.AuthContext) error

func (f PolicyFunc) IsAllowed(r *http.Request, ctx handler.AuthContext) error {
	return f(r, ctx)
}
