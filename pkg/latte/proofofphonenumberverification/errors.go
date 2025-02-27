package proofofphonenumberverification

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidConfiguration = apierrors.InternalError.WithReason("InvalidConfiguration")
