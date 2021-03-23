package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	PreSignup event.Type = "pre_signup"
)

type PreSignupBlockingEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (e *PreSignupBlockingEvent) BlockingEventType() event.Type {
	return PreSignup
}

func (e *PreSignupBlockingEvent) UserID() string {
	return e.User.ID
}

var _ event.BlockingPayload = &PreSignupBlockingEvent{}
