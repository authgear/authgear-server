package oob

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidOOBCode = apierrors.Forbidden.WithReason("InvalidOOBCode")

var ErrCodeNotFound = InvalidOOBCode.NewWithCause("oob code is expired or invalid", apierrors.StringCause("CodeNotFound"))
