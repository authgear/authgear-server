package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var AccessControlViolated = apierrors.Forbidden.WithReason("AccessControlViolated")
var InvariantViolated = apierrors.Invalid.WithReason("InvariantViolated")
