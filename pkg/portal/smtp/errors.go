package smtp

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var (
	SMTPTestFailed = apierrors.InternalError.WithReason("SMTPTestFailed").SkipLoggingToExternalService()
)
