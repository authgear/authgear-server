package slice

func FlatMap[A any, B any](as []A, mapper func(A) []B) []B {
	bs := []B{}
	for _, a := range as {
		blist := mapper(a)
		bs = append(bs, blist...)
	}
	return bs
}
