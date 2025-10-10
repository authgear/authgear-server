package hook

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var HookDisallowed = apierrors.Forbidden.WithReason("HookDisallowed")
var HookDeliveryTimeout = apierrors.InternalError.WithReason("HookDeliveryTimeout").SkipLoggingToExternalService()
var HookInvalidResponse = apierrors.InternalError.WithReason("HookInvalidResponse").SkipLoggingToExternalService()

var HookDeliveryUnknownFailure = apierrors.InternalError.WithReason("HookDeliveryUnknownFailure").SkipLoggingToExternalService()

var DenoRunError = apierrors.BadRequest.WithReason("DenoRunError")

var DenoCheckError = apierrors.Invalid.WithReason("DenoCheckError")

type OperationDisallowedItem struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

func newErrorOperationDisallowed(eventType string, items []OperationDisallowedItem) error {
	// These are not causes. Causes are pre-defined, and reasons are provided by hook handlers.
	return HookDisallowed.NewWithInfo(
		"disallowed by hook event handler",
		map[string]interface{}{
			"event_type": eventType,
			"reasons":    items,
		},
	)
}
