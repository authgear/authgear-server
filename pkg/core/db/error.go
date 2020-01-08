package db

import "github.com/skygeario/skygear-server/pkg/core/errors"

var ErrWriteConflict = errors.New("concurrent write conflict occurred")
