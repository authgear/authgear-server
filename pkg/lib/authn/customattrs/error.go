package customattrs

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var AccessControlViolated = apierrors.Forbidden.WithReason("AccessControlViolated")
