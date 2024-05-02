package hook

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var WebHookDisallowed = apierrors.Forbidden.WithReason("WebHookDisallowed")

var WebHookDeliveryTimeout = apierrors.InternalError.WithReason("WebHookDeliveryTimeout")
var WebHookInvalidResponse = apierrors.InternalError.WithReason("WebHookInvalidResponse")

var DenoRunError = apierrors.BadRequest.WithReason("DenoRunError")

var DenoCheckError = apierrors.Invalid.WithReason("DenoCheckError")

func init() {
	apierrors.SkipLoggingForKinds[WebHookInvalidResponse] = true
}

type OperationDisallowedItem struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

func newErrorOperationDisallowed(eventType string, items []OperationDisallowedItem) error {
	// These are not causes. Causes are pre-defined, and reasons are provided by hook handlers.
	return WebHookDisallowed.NewWithInfo(
		"disallowed by web-hook event handler",
		map[string]interface{}{
			"event_type": eventType,
			"reasons":    items,
		},
	)
}
