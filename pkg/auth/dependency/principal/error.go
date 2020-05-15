package principal

import (
	"errors"
)

var ErrNotFound = errors.New("principal not found")
var ErrAlreadyExists = errors.New("principal already exists")
var ErrMultipleResultsFound = errors.New("multiple principals found")
