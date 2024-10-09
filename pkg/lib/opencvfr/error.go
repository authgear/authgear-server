package opencvfr

import "fmt"

var ErrFaceNotFound = fmt.Errorf("face not found")
var ErrFaceNotMatch = fmt.Errorf("face does not match target person")
var ErrFaceLivenessLow = fmt.Errorf("face liveness score too low")
