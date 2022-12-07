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

func (e *UserPreScheduleDeletionBlockingEventPayload) ApplyMutations(mutations event.Mutations) bool {
	user, mutated := ApplyUserMutations(e.UserModel, mutations.User)
	if mutated {
		e.UserModel = user
		return true
	}

	return false
}

func (e *UserPreScheduleDeletionBlockingEventPayload) PerformEffects(ctx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, userID, userMutations)
}

var _ event.BlockingPayload = &UserPreScheduleDeletionBlockingEventPayload{}
