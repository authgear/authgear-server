package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type inputTriggerOOB struct {
	AuthenticatorIndex int
	AuthenticatorType  string
}

var _ interaction.Input = &inputTriggerOOB{}
var _ nodes.InputAuthenticationOOBTrigger = &inputTriggerOOB{}

func (i *inputTriggerOOB) GetOOBAuthenticatorIndex() int   { return i.AuthenticatorIndex }
func (i *inputTriggerOOB) GetOOBAuthenticatorType() string { return i.AuthenticatorType }
func (*inputTriggerOOB) IsInteractive() bool               { return false }

type inputSelectTOTP struct{}

var _ interaction.Input = &inputSelectTOTP{}
var _ nodes.InputCreateAuthenticatorTOTPSetup = &inputSelectTOTP{}

func (i *inputSelectTOTP) SetupTOTP()        {}
func (*inputSelectTOTP) IsInteractive() bool { return false }

type inputAuthDeviceToken struct {
	DeviceToken string
}

var _ interaction.Input = &inputAuthDeviceToken{}
var _ nodes.InputUseDeviceToken = &inputAuthDeviceToken{}

func (i *inputAuthDeviceToken) GetDeviceToken() string { return i.DeviceToken }
func (*inputAuthDeviceToken) IsInteractive() bool      { return false }
