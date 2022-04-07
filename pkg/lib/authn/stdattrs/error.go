package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var AccessControlViolated = apierrors.Forbidden.WithReason("AccessControlViolated")
var StandardAttributesEmailRequired = apierrors.BadRequest.WithReason("StandardAttributesEmailRequired")
