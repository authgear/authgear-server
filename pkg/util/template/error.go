package template

import (
	"errors"
)

var ErrLimitReached = errors.New("template: rendered template is too large")
var ErrNotFound = errors.New("requested template not found")
