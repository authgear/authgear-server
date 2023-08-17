package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
}

type inputTakeLoginID interface {
	GetLoginID() string
}

type inputTakeOOBOTPChannel interface {
	GetChannel() model.AuthenticatorOOBChannel
}

type inputNodeVerifyClaim interface {
	IsCode() bool
	IsResend() bool
	IsCheck() bool
	GetCode() string
}
