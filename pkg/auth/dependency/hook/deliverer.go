package hook

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Deliverer interface {
	DeliverBeforeEvent(event *event.Event, user *model.User) error
	DeliverNonBeforeEvent(event *event.Event, timeout time.Duration) error
}
