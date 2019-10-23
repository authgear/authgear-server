package sso

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var SSOFailed = skyerr.Unauthorized.WithReason("SSOFailed")

type ssoFailCause string

const (
	SSOUnauthorized ssoFailCause = "Unauthorized"
	NetworkFailed   ssoFailCause = "NetworkFailed"
	InvalidParams   ssoFailCause = "InvalidParams"
)

func NewSSOFailed(reason ssoFailCause, msg string) error {
	return SSOFailed.NewWithDetails(msg, skyerr.Details{"cause": skyerr.APIErrorString(reason)})
}
