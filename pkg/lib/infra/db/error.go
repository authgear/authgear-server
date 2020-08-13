package db

import "errors"

var ErrWriteConflict = errors.New("concurrent write conflict occurred")
