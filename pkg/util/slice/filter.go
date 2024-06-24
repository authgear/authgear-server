package slice

func Filter[A any](as []A, pred func(A) bool) []A {
	result := []A{}
	for _, a := range as {
		aa := a
		if pred(a) {
			result = append(result, aa)
		}
	}
	return result
}
