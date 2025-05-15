package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InactiveOAuthProvider = apierrors.Invalid.WithReason("InactiveOAuthProvider")
