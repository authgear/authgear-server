package hook

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
)

type InvalidEventPayload struct {
	payload event.Payload
}

func (err InvalidEventPayload) Error() string {
	return fmt.Sprintf("invalid event payload: %T", err.payload)
}

type DeliveryTimeout struct{}

func (err DeliveryTimeout) Error() string {
	return "web-hook event delivery timed out"
}

type DeliveryFailed struct {
	inner error
}

func (err DeliveryFailed) Error() string {
	return fmt.Sprintf("web-hook event delivery failed: %v", err.inner)
}

var DeliveryFailedInvalidStatusCode = DeliveryFailed{
	inner: fmt.Errorf("invalid status code"),
}

type OperationDisallowed struct {
	Items []OperationDisallowedItem
}
type OperationDisallowedItem struct {
	Reason string      `json:"reason"`
	Data   interface{} `json:"data,omitempty"`
}

func (err OperationDisallowed) Error() string {
	return "disallowed by web-hook event handler"
}

type MutationFailed struct {
	inner error
}

func (err MutationFailed) Error() string {
	return fmt.Sprintf("web-hook mutation failed: %v", err.inner)
}
