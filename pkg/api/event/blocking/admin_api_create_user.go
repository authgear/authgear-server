package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPICreateUser event.Type = "admin_api_create_user"
)

type AdminAPICreateUserBlockingEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (e *AdminAPICreateUserBlockingEvent) BlockingEventType() event.Type {
	return AdminAPICreateUser
}

func (e *AdminAPICreateUserBlockingEvent) UserID() string {
	return e.User.ID
}

var _ event.BlockingPayload = &AdminAPICreateUserBlockingEvent{}
