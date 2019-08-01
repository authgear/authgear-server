package hook

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type invalidEventPayload struct {
	payload event.Payload
}

func (err invalidEventPayload) Error() string {
	return fmt.Sprintf("invalid event payload: %T", err.payload)
}

func newErrorDeliveryTimeout() error {
	return skyerr.NewError(skyerr.WebHookTimeOut, "web-hook event delivery timed out")
}

func newErrorDeliveryFailed(inner error) error {
	return skyerr.NewErrorf(skyerr.WebHookFailed, "web-hook event delivery failed: %v", inner)
}

func newErrorDeliveryInvalidStatusCode() error {
	return skyerr.NewError(skyerr.WebHookFailed, "invalid status code")
}

type OperationDisallowedItem struct {
	Reason string      `json:"reason"`
	Data   interface{} `json:"data,omitempty"`
}

func newErrorOperationDisallowed(items []OperationDisallowedItem) error {
	return skyerr.NewErrorWithInfo(
		skyerr.PermissionDenied,
		"disallowed by web-hook event handler",
		map[string]interface{}{"errors": items},
	)
}

func newErrorMutationFailed(inner error) error {
	return skyerr.NewErrorf(skyerr.WebHookFailed, "web-hook mutation failed: %v", inner)
}
