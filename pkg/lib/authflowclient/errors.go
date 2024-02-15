package authflowclient

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowNotFound").New("flow not found")
