package hook

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

var WebHookDisallowed = apierrors.Forbidden.WithReason("WebHookDisallowed")

var errDeliveryTimeout = errors.New("web-hook event delivery timed out")
var errDeliveryInvalidStatusCode = errors.New("invalid status code")

func newErrorDeliveryFailed(inner error) error {
	return fmt.Errorf("web-hook event delivery failed: %w", inner)
}

type OperationDisallowedItem struct {
	Reason string      `json:"reason"`
	Data   interface{} `json:"data,omitempty"`
}

func newErrorOperationDisallowed(items []OperationDisallowedItem) error {
	// NOTE(error): These are not causes. Causes are pre-defined,
	// and reasons are provided by hook handlers.
	return WebHookDisallowed.NewWithInfo(
		"disallowed by web-hook event handler",
		map[string]interface{}{"reasons": items},
	)
}
