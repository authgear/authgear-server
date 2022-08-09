package facade

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvariantViolated = apierrors.Invalid.WithReason("InvariantViolated")

func NewInvariantViolated(cause string, msg string, data map[string]interface{}) error {
	return InvariantViolated.NewWithCause(
		msg,
		apierrors.MapCause{
			CauseKind: cause,
			Data:      data,
		},
	)
}
