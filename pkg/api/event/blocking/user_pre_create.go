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

func (e *UserPreCreateBlockingEventPayload) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	user, mutated := ApplyMutations(e.UserModel, mutations)
	if mutated {
		copied := *e
		copied.UserModel = user
		return &copied, true
	}

	return e, false
}

func (e *UserPreCreateBlockingEventPayload) GenerateFullMutations() event.Mutations {
	return GenerateFullMutations(e.UserModel)
}

var _ event.BlockingPayload = &UserPreCreateBlockingEventPayload{}
