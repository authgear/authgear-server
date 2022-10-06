package oauth

type AuthorizationFilter interface {
	Keep(authz *Authorization) bool
}

type AuthorizationFilterFunc func(a *Authorization) bool

func (f AuthorizationFilterFunc) Keep(a *Authorization) bool {
	return f(a)
}
