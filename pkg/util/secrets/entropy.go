package secrets

import (
	"math"
)

// ShannonEntropy is borrowed from
// https://github.com/Xe/x/blob/v1.2.3/entropy/shannon.go.
func ShannonEntropy(s string) (bits int) {
	numberOfOccurrence := make(map[rune]int)
	for _, r := range s {
		numberOfOccurrence[r]++
	}

	var sum float64
	for _, v := range numberOfOccurrence {
		probability := float64(v) / float64(len(s))
		sum += probability * math.Log2(probability)
	}

	bitsPerRune := int(math.Ceil(sum * -1))
	bits = bitsPerRune * len(s)
	return
}
