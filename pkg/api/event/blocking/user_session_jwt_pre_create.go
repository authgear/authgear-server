package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSessionJWTPreCreate event.Type = "user.session.jwt.pre_create"
)

type UserSessionJWTPreCreateBlockingEventPayload struct {
	UserRef   model.UserRef          `json:"-" resolve:"user"`
	UserModel model.User             `json:"user"`
	Payload   map[string]interface{} `json:"payload"`
}

func (e *UserSessionJWTPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return UserSessionJWTPreCreate
}

func (e *UserSessionJWTPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSessionJWTPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserSessionJWTPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserSessionJWTPreCreateBlockingEventPayload) ApplyMutations(mutations event.Mutations) bool {
	if mutations.JWT.Payload != nil {
		e.Payload = mutations.JWT.Payload
		return true
	}

	return false
}

func (e *UserSessionJWTPreCreateBlockingEventPayload) PerformEffects(ctx event.MutationsEffectContext) error {
	return nil
}

var _ event.BlockingPayload = &UserSessionJWTPreCreateBlockingEventPayload{}
