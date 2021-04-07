package hook

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var WebHookDisallowed = apierrors.Forbidden.WithReason("WebHookDisallowed")

var errDeliveryTimeout = errors.New("web-hook event delivery timed out")
var errDeliveryInvalidStatusCode = errors.New("invalid status code")

func newErrorDeliveryFailed(inner error) error {
	return fmt.Errorf("web-hook event delivery failed: %w", inner)
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
