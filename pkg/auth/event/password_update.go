package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforePasswordUpdate Type = "before_password_update"
	AfterPasswordUpdate  Type = "after_password_udpate"
)

type PasswordUpdateReason string

const (
	PasswordUpdateReasonChangePassword = "change-password"
	PasswordUpdateReasonResetPassword  = "reset-password"
	PasswordUpdateReasonAdministrative = "administrative"
)

type PasswordUpdateEvent struct {
	Reason PasswordUpdateReason `json:"reason"`
	User   *model.User          `json:"user"`
}

func (PasswordUpdateEvent) Version() int32 {
	return 1
}
