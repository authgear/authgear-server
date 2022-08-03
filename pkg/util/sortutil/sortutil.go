package sortutil

type LessFunc func(i int, j int) bool

func (f LessFunc) AndThen(g LessFunc) LessFunc {
	return LessFunc(func(i int, j int) bool {
		fij := f(i, j)
		fji := f(j, i)

		switch {
		case fij && !fji:
			return true
		case !fij && fji:
			return false
		default:
			return g(i, j)
		}
	})
}
