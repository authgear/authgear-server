package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Deliverer interface {
	DeliverBeforeEvent(event *event.Event, user *model.User) error
}
