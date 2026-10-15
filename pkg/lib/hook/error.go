package hook

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
)

var HookDisallowed = apierrors.Forbidden.WithReason("HookDisallowed")
var HookDeliveryTimeout = apierrors.InternalError.WithReason("HookDeliveryTimeout").SkipLoggingToExternalService()
var HookInvalidResponse = apierrors.InternalError.WithReason("HookInvalidResponse").SkipLoggingToExternalService()

var HookDeliveryUnknownFailure = apierrors.InternalError.WithReason("HookDeliveryUnknownFailure").SkipLoggingToExternalService()

var DenoRunError = apierrors.BadRequest.WithReason("DenoRunError")

var DenoCheckError = apierrors.Invalid.WithReason("DenoCheckError")

type OperationDisallowedItem struct {
	Title   string                      `json:"title"`
	Reason  string                      `json:"reason"`
	Reasons []OperationDisallowedReason `json:"reasons,omitempty"`
}

type OperationDisallowedReason struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

func toOperationDisallowedReasons(reasons []event.HookResponseReason) []OperationDisallowedReason {
	var out []OperationDisallowedReason
	for _, r := range reasons {
		item := OperationDisallowedReason{
			Type:    string(r.Type),
			Message: r.Message,
		}
		if r.Field != nil {
			item.Key = r.Field.Key()
		}
		out = append(out, item)
	}
	return out
}

func newErrorOperationDisallowed(eventType string, items []OperationDisallowedItem) error {
	// These are not causes. Causes are pre-defined, and reasons are provided by hook handlers.
	return HookDisallowed.NewWithInfo(
		"disallowed by hook event handler",
		map[string]any{
			"event_type": eventType,
			"reasons":    items,
		},
	)
}
