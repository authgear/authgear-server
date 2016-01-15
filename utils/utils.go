package utils

func StrSliceWithout(slice []string, without []string) []string {
	newSlice := []string{}

	for _, c := range slice {
		if pos := strAt(without, c); pos == -1 {
			newSlice = append(newSlice, c)
		}
	}
	return newSlice
}

func strAt(slice []string, str string) int {
	for pos, s := range slice {
		if s == str {
			return pos
		}
	}
	return -1
}
