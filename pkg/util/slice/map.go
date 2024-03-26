package slice

func Map[A any, B any](as []A, mapper func(A) B) []B {
	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = mapper(a)
	}
	return bs
}
