package strings

func GetFirstN(str string, n int) string {
	if n < 0 {
		panic("n must be >= 0")
	}
	v := []rune(str)
	if n >= len(v) {
		return str
	}
	return string(v[:n])
}
