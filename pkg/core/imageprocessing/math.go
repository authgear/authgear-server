package imageprocessing

import (
	"math"
)

func ratio(x int, y int) float64 {
	return float64(x) / float64(y)
}

func roundFloat(f float64) int {
	return int(math.Round(f))
}
