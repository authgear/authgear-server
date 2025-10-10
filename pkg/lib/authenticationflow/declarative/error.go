package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowFlowNotFound").New("flow not found")

var InvalidTargetStep = apierrors.InternalError.WithReason("AuthenticationFlowInvalidTargetStep")
var InvalidFlowConfig = apierrors.InternalError.WithReason("AuthenticationFlowInvalidFlowConfig")

var ErrDifferentUserID = authenticationflow.ErrDifferentUserID
var ErrNoUserID = authenticationflow.ErrNoUserID

var ErrNoPublicSignup = apierrors.Forbidden.
	WithReason("AuthenticationFlowNoPublicSignup").
	SkipLoggingToExternalService().
	New("public signup is disabled")
