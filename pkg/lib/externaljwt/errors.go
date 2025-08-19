package externaljwt

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrExternalJWTInvalidJWT = apierrors.BadRequest.WithReason("ExternalJWTInvalidJWT")
var ErrExternalJWTFailedToFetchJWKs = apierrors.InternalError.WithReason("ExternalJWTFailedToFetchJWKs").SkipLoggingToExternalService()
var ErrExternalJWTInvalidClaim = apierrors.BadRequest.WithReason("ExternalJWTInvalidClaim")
