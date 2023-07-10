package slice

func FindIndex[T any](in []T, findFn func(item T) bool) int {
	result := -1
	for idx, item := range in {
		if findFn(item) {
			result = idx
			break
		}
	}
	return result
}
