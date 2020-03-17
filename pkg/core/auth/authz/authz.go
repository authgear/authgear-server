package authz

import (
	"net/http"
)

type PolicyProvider interface {
	ProvideAuthzPolicy() Policy
}

type Policy interface {
	IsAllowed(r *http.Request) error
}

type PolicyFunc func(r *http.Request) error

func (f PolicyFunc) IsAllowed(r *http.Request) error {
	return f(r)
}
