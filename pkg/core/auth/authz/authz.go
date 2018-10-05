package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
)

type PolicyProvider interface {
	ProvideAuthzPolicy() Policy
}

type Policy interface {
	IsAllowed(r *http.Request, ctx context.AuthContext) error
}

type PolicyFunc func(r *http.Request, ctx context.AuthContext) error

func (f PolicyFunc) IsAllowed(r *http.Request, ctx context.AuthContext) error {
	return f(r, ctx)
}
