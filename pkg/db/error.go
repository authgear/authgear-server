package db

import "github.com/authgear/authgear-server/pkg/core/errors"

var ErrWriteConflict = errors.New("concurrent write conflict occurred")
