package setutil

import (
	"cmp"
	"slices"
)

type Set[T cmp.Ordered] map[T]struct{}

func Identity[T any](t T) T {
	return t
}

func NewSetFromSlice[A any, B cmp.Ordered](slice []A, f func(a A) B) Set[B] {
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

func SetToSlice[A any, B cmp.Ordered](slice []A, set Set[B], f func(a A) B) []A {
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
	if s == nil {
		return []T{}
	}
	keys := []T{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, cmp.Compare[T])
	return keys
}

func (s *Set[T]) Add(key T) {
	if (*s) == nil {
		*s = Set[T]{}
	}
	if _, ok := (*s)[key]; ok {
		return
	}
	(*s)[key] = struct{}{}
}

func (s *Set[T]) Delete(key T) {
	if (*s) == nil {
		*s = Set[T]{}
	}
	delete(*s, key)
}

func (s Set[T]) Has(key T) bool {
	if s == nil {
		return false
	}
	if _, ok := s[key]; ok {
		return true
	}
	return false
}

func (s *Set[T]) Merge(other Set[T]) Set[T] {
	result := Set[T]{}
	if (*s) != nil {
		for _, k := range s.Keys() {
			result.Add(k)
		}
	}
	for _, k := range other.Keys() {
		result.Add(k)
	}
	return result
}
