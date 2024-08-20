package loginid

import (
	"errors"
)

var ErrValidate = errors.New("ValidateError")
var ErrNormalize = errors.New("NormalizationError")
