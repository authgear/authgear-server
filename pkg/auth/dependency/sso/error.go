package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var SSOFailed = skyerr.Unauthorized.WithReason("SSOFailed")

type ssoFailCause string

const (
	SSOUnauthorized ssoFailCause = "Unauthorized"
	NetworkFailed   ssoFailCause = "NetworkFailed"
	InvalidParams   ssoFailCause = "InvalidParams"
	AlreadyLinked   ssoFailCause = "AlreadyLinked"
)

func NewSSOFailed(reason ssoFailCause, msg string) error {
	return SSOFailed.NewWithInfo(msg, skyerr.Details{"cause": reason})
}
