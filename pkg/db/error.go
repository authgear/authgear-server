package db

import "github.com/authgear/authgear-server/pkg/util/errors"

var ErrWriteConflict = errors.New("concurrent write conflict occurred")
