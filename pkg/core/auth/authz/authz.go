package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type PolicyProvider interface {
	ProvideAuthzPolicy() Policy
}

type Policy interface {
	IsAllowed(r *http.Request, ctx auth.ContextGetter) error
}

type PolicyFunc func(r *http.Request, ctx auth.ContextGetter) error

func (f PolicyFunc) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	return f(r, ctx)
}
