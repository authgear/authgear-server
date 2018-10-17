package handler

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var timeNow = func() time.Time { return time.Now().UTC() }

type serializedError struct {
	id  string
	err skyerr.Error
}

func newSerializedError(id string, err skyerr.Error) serializedError {
	return serializedError{
		id:  id,
		err: err,
	}
}
