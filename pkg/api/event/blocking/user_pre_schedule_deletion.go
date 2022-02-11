package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreScheduleDeletion event.Type = "user.pre_schedule_deletion"
)

type UserPreScheduleDeletionBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserPreScheduleDeletionBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreScheduleDeletion
}

func (e *UserPreScheduleDeletionBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreScheduleDeletionBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreScheduleDeletionBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserPreScheduleDeletionBlockingEventPayload) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	user, mutated := ApplyMutations(e.UserModel, mutations)
	if mutated {
		copied := *e
		copied.UserModel = user
		return &copied, true
	}

	return e, false
}

func (e *UserPreScheduleDeletionBlockingEventPayload) GenerateFullMutations() event.Mutations {
	return GenerateFullMutations(e.UserModel)
}

var _ event.BlockingPayload = &UserPreScheduleDeletionBlockingEventPayload{}
