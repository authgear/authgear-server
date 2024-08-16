package setutil

type Set[T comparable] map[T]struct{}

func Identity[T any](t T) T {
	return t
}

func NewSetFromSlice[A any, B comparable](slice []A, f func(a A) B) Set[B] {
	out := make(Set[B])
	for _, a := range slice {
		b := f(a)
		out[b] = struct{}{}
	}
	return out
}

func (s Set[T]) Subtract(that Set[T]) Set[T] {
	out := make(Set[T])
	for inThis := range s {
		_, ok := that[inThis]
		if !ok {
			out[inThis] = struct{}{}
		}
	}
	return out
}

func SetToSlice[A any, B comparable](slice []A, set Set[B], f func(a A) B) []A {
	var out []A
	for _, a := range slice {
		b := f(a)
		_, ok := set[b]
		if ok {
			out = append(out, a)
		}
	}
	return out
}

func (s Set[T]) Keys() []T {
	keys := []T{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s Set[T]) Add(key T) {
	if _, ok := s[key]; ok {
		return
	}
	s[key] = struct{}{}
}

func (s Set[T]) Has(key T) bool {
	if _, ok := s[key]; ok {
		return true
	}
	return false
}
