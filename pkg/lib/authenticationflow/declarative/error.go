package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowFlowNotFound").New("flow not found")

var InvalidTargetStep = apierrors.InternalError.WithReason("AuthenticationFlowInvalidTargetStep")
var InvalidFlowConfig = apierrors.InternalError.WithReason("AuthenticationFlowInvalidFlowConfig")

var ErrDifferentUserID = apierrors.BadRequest.WithReason("AuthenticationFlowDifferentUserID").New("different user ID")
var ErrNoUserID = apierrors.BadRequest.WithReason("AuthenticationFlowNoUserID").New("no user ID")

var ErrNoPublicSignup = apierrors.Forbidden.WithReason("AuthenticationFlowNoPublicSignup").New("public signup is disabled")
