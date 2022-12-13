package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreCreate event.Type = "user.pre_create"
)

type UserPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	OAuthState string           `json:"-"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreCreate
}

func (e *UserPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {
	if e.OAuthState != "" {
		ctx.OAuth = &event.OAuthContext{
			State: e.OAuthState,
		}
	}
}

func (e *UserPreCreateBlockingEventPayload) ApplyMutations(mutations event.Mutations) bool {
	user, mutated := ApplyUserMutations(e.UserModel, mutations.User)
	if mutated {
		e.UserModel = user
		return true
	}

	return false
}

func (e *UserPreCreateBlockingEventPayload) PerformEffects(ctx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, userID, userMutations)
}

var _ event.BlockingPayload = &UserPreCreateBlockingEventPayload{}
