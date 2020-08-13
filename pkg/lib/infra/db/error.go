package db

import "github.com/authgear/authgear-server/pkg/util/errorutil"

var ErrWriteConflict = errorutil.New("concurrent write conflict occurred")
