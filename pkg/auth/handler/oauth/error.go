package oauth

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidGrant = apierrors.Forbidden.WithReason("InvalidGrant")
