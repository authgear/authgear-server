package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreCreate event.Type = "user.pre_create"
)

type UserPreCreateBlockingEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	OAuthState string           `json:"-"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserPreCreateBlockingEvent) BlockingEventType() event.Type {
	return UserPreCreate
}

func (e *UserPreCreateBlockingEvent) UserID() string {
	return e.User.ID
}

func (e *UserPreCreateBlockingEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserPreCreateBlockingEvent) FillContext(ctx *event.Context) {
	if e.OAuthState != "" {
		ctx.OAuth = &event.OAuthContext{
			State: e.OAuthState,
		}
	}
}

var _ event.BlockingPayload = &UserPreCreateBlockingEvent{}
