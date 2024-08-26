package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OIDCJWTPreCreate event.Type = "oidc.jwt.pre_create"
)

type OIDCJWT struct {
	Payload map[string]interface{} `json:"payload"`
}

type OIDCJWTPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	JWT        OIDCJWT          `json:"jwt"`
}

func (e *OIDCJWTPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return OIDCJWTPreCreate
}

func (e *OIDCJWTPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *OIDCJWTPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *OIDCJWTPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *OIDCJWTPreCreateBlockingEventPayload) ApplyMutations(mutations event.Mutations) bool {
	if mutations.JWT.Payload != nil {
		e.JWT.Payload = mutations.JWT.Payload
		return true
	}

	return false
}

func (e *OIDCJWTPreCreateBlockingEventPayload) PerformEffects(ctx event.MutationsEffectContext) error {
	return nil
}

var _ event.BlockingPayload = &OIDCJWTPreCreateBlockingEventPayload{}
