package slice

// ExceptStrings return a new slice that without the element appears in the
// second slice.
func ExceptStrings(slice []string, except []string) []string {
	newSlice := []string{}

	for _, c := range slice {
		if pos := strAt(except, c); pos == -1 {
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

// ContainsString determine whether the input slice contains the specified string.
func ContainsString(in []string, elem string) bool {
	for i := 0; i < len(in); i++ {
		if in[i] == elem {
			return true
		}
	}

	return false
}
