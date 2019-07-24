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
