package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("WorkflowConfigFlowNotFound").New("flow not found")

var InvalidIdentificationMethod = apierrors.BadRequest.WithReason("workflowConfigInvalidIdentificationMethod")
