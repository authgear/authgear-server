package cmdinternal

import (
	"fmt"
	"strings"
)

func mapGet[T any](m map[string]any, path ...string) (result T, found bool) {
	if len(path) == 0 {
		result, found = any(m).(T)
		return
	}

	current := m
	last := len(path) - 1

	for i, key := range path {
		value, ok := current[key]
		if !ok {
			found = false
			return
		}
		if i == last {
			result, found = value.(T)
			return
		}

		current, ok = value.(map[string]any)
		if !ok {
			found = false
			return
		}
	}
	return
}

func mapSet[T any](m map[string]any, value T, path ...string) {
	current := m
	last := len(path) - 1

	for i, key := range path {
		if i == last {
			current[key] = value
			return
		}

		m, ok := current[key]
		if !ok {
			m = map[string]any{}
			current[key] = m
		}
		current, ok = m.(map[string]any)
		if !ok {
			panic(fmt.Sprintf("unexpected value type for for path %q: %T", strings.Join(path, "."), m))
		}
	}
}

func mapSetIfNotFound[T any](m map[string]any, value T, path ...string) {
	_, found := mapGet[T](m, path...)
	if !found {
		mapSet(m, value, path...)
	}
}

func mapDelete(m map[string]any, path ...string) {
	m, found := mapGet[map[string]any](m, path[:len(path)-1]...)
	if found {
		delete(m, path[len(path)-1])
	}
}
