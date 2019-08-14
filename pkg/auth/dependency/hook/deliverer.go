package hook

import (
	"net/url"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Deliverer interface {
	WillDeliver(eventType event.Type) bool
	DeliverBeforeEvent(baseURL *url.URL, event *event.Event, user *model.User) error
	DeliverNonBeforeEvent(baseURL *url.URL, event *event.Event, timeout time.Duration) error
}
