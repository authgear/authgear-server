package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Mutator interface {
	New(event *event.Event, user *model.User) Mutator
	Add(event.Mutations)
	Apply() error
}
