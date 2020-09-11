package db

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrWriteConflict = apierrors.NewDataRace("concurrent write conflict occurred")
