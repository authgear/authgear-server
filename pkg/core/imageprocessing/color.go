package imageprocessing

// Color stores 24-bit color.
type Color struct {
	// R is in range [0,255].
	R int
	// G is in range [0,255].
	G int
	// B is in range [0,255].
	B int
}

// ColorWhite is white.
var ColorWhite = Color{
	R: 255,
	G: 255,
	B: 255,
}
