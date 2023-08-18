package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
}

type inputTakeAuthenticationMethod interface {
	GetAuthenticationMethod() config.WorkflowAuthenticationMethod
}

type inputTakeLoginID interface {
	GetLoginID() string
}

type inputTakeOOBOTPChannel interface {
	GetChannel() model.AuthenticatorOOBChannel
}

type inputTakeOOBOTPTarget interface {
	GetTarget() string
}

type inputTakeNewPassword interface {
	GetNewPassword() string
}

type inputNodeVerifyClaim interface {
	IsCode() bool
	IsResend() bool
	IsCheck() bool
	GetCode() string
}

type inputSetupTOTP interface {
	GetCode() string
	GetDisplayName() string
}

type inputConfirmRecoveryCode interface {
	ConfirmRecoveryCode()
}

type inputFillUserProfile interface {
	GetAttributes() []attrs.T
}

type inputConfirmTerminateOtherSessions interface {
	ConfirmTerminateOtherSessions()
}
