package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowFlowNotFound").New("flow not found")
var ErrStepNotFound = apierrors.NotFound.WithReason("AuthenticationFlowStepNotFound").New("step not found")

var InvalidTargetStep = apierrors.InternalError.WithReason("AuthenticationFlowInvalidTargetStep")

var ErrDifferentUserID = apierrors.BadRequest.WithReason("AuthenticationFlowDifferentUserID").New("different user ID")
var ErrNoUserID = apierrors.BadRequest.WithReason("AuthenticationFlowNoUserID").New("no user ID")
