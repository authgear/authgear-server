package oauth

type Filter interface {
	Keep(authz *Authorization) bool
}

type FilterFunc func(a *Authorization) bool

func (f FilterFunc) Keep(a *Authorization) bool {
	return f(a)
}
