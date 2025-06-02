package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserProfilePreUpdate event.Type = "user.profile.pre_update"
)

type UserProfilePreUpdateBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserProfilePreUpdateBlockingEventPayload) BlockingEventType() event.Type {
	return UserProfilePreUpdate
}

func (e *UserProfilePreUpdateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserProfilePreUpdateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserProfilePreUpdateBlockingEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserProfilePreUpdateBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	user, mutated := ApplyUserMutations(e.UserModel, response.Mutations.User)
	if mutated {
		e.UserModel = user
	}
	return event.ApplyHookResponseResult{UserMutationsEverApplied: mutated}
}

func (e *UserProfilePreUpdateBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, effectCtx, userID, userMutations)
}

var _ event.BlockingPayload = &UserProfilePreUpdateBlockingEventPayload{}
