package slice

func Cast[A any, B any](as []A) []B {
	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = any(a).(B)
	}
	return bs
}
