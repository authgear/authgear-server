package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFlowNotFound = apierrors.NotFound.WithReason("WorkflowConfigFlowNotFound").New("flow not found")
var ErrStepNotFound = apierrors.NotFound.WithReason("WorkflowConfigStepNotFound").New("step not found")

var InvalidIdentificationMethod = apierrors.BadRequest.WithReason("WorkflowConfigInvalidIdentificationMethod")
var InvalidVerifyTarget = apierrors.InternalError.WithReason("WorkflowConfigInvalidVerifyTarget")
