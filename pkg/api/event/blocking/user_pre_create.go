package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreCreate event.Type = "user.pre_create"
)

type UserPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-"`
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

func (e *UserPreCreateBlockingEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {
	if e.OAuthState != "" {
		ctx.OAuth = &event.OAuthContext{
			State: e.OAuthState,
		}
	}
}

func (e *UserPreCreateBlockingEventPayload) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	if mutations.User.StandardAttributes != nil {
		copied := *e
		copied.UserModel.StandardAttributes = mutations.User.StandardAttributes
		return &copied, true
	}

	return e, false
}

func (e *UserPreCreateBlockingEventPayload) GenerateFullMutations() event.Mutations {
	return event.Mutations{
		User: event.UserMutations{
			StandardAttributes: e.UserModel.StandardAttributes,
		},
	}
}

var _ event.BlockingPayload = &UserPreCreateBlockingEventPayload{}
