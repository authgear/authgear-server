package sso

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var SSOFailed = apierrors.Unauthorized.WithReason("SSOFailed")

type ssoFailCause string

const (
	SSOUnauthorized     ssoFailCause = "Unauthorized"
	NetworkFailed       ssoFailCause = "NetworkFailed"
	InvalidParams       ssoFailCause = "InvalidParams"
	AlreadyLinked       ssoFailCause = "AlreadyLinked"
	InvalidCodeVerifier ssoFailCause = "InvalidCodeVerifier"
)

func NewSSOFailed(cause ssoFailCause, msg string) error {
	return SSOFailed.NewWithCause(msg, apierrors.StringCause(cause))
}
