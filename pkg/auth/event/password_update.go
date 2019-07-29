package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforePasswordUpdate Type = "before_password_update"
	AfterPasswordUpdate  Type = "after_password_udpate"
)

type PasswordUpdateReason string

const (
	PasswordUpdateReasonChangePassword = "change_password"
	PasswordUpdateReasonResetPassword  = "reset_password"
	PasswordUpdateReasonAdministrative = "administrative"
)

type PasswordUpdateEvent struct {
	Reason PasswordUpdateReason `json:"reason"`
	User   model.User           `json:"user"`
}

func (PasswordUpdateEvent) BeforeEventType() Type {
	return BeforePasswordUpdate
}

func (PasswordUpdateEvent) AfterEventType() Type {
	return AfterPasswordUpdate
}

func (event PasswordUpdateEvent) ApplyingMutations(mutations Mutations) UserAwarePayload {
	// user object in this event is a snapshot before operation, so mutations are not applied
	return event
}

func (event PasswordUpdateEvent) UserID() string {
	return event.User.ID
}
