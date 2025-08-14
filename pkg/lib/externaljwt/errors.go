package externaljwt

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidExternalJWT = apierrors.BadRequest.WithReason("InvalidExternalJWT")
var ErrFailedToFetchJWKS = apierrors.InternalError.WithReason("FailedToFetchJWKS").SkipLoggingToExternalService()
