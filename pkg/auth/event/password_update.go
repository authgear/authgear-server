package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforePasswordUpdate Type = "before_password_update"
	AfterPasswordUpdate  Type = "after_password_udpate"
)

const PasswordUpdateEventVersion int32 = 1

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
